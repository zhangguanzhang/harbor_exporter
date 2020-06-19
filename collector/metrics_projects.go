package collector

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
)

// check interface
var _ Scraper = ScrapeProjects{}

const (
	projectsUrl  = "/projects"
)

var (
	projectsRefInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ref_work", "projects"),
		"test the projects api ref work status(0 for error, 1 for success).",
		[]string{"ref", "method"}, nil,
	)
	reposRefInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ref_work", "repos"),
		"test the repos api ref work status(0 for error, 1 for success).",
		[]string{"ref", "method"}, nil,
	)

)

type projectsJson struct {
	ProjectID int `json:"project_id"`
}

type metadataJson struct {
	Public string `json:"public"`
}


type ScrapeProjects struct{}

// Name of the Scraper. Should be unique.
func (ScrapeProjects) Name() string {
	return "projects"
}

// Help describes the role of the Scraper.
func (ScrapeProjects) Help() string {
	return "Collect the projects and repos api work"
}

// Scrape collects data from client and sends it over channel as prometheus metric.
func (ScrapeProjects) Scrape(client *HarborClient, ch chan<- prometheus.Metric) error {
	var (
		id int
		err error
	)

	id, err = projects(client, ch)
	if err != nil {return err}

	err = projectsLogs(id, client, ch)
	if err != nil {return err}

	err = projectsMetadata(id, client, ch)
	if err != nil {return err}

	err = projectsMembers(id, client, ch)
	if err != nil {return err}

	err = reposQuery(id, client, ch)
	if err != nil {return err}

	err = reposTop(client, ch)
	if err != nil {return err}

	return nil
}


func projects(client *HarborClient, ch chan<- prometheus.Metric) (int, error) {
	var data []projectsJson
	url := projectsUrl + "?page_size=1"
	body, err := client.request(url)
	if err != nil {
		return 0, err
	}

	if err = json.Unmarshal(body, &data); err != nil {
		return 0, err
	}

	if len(data) != 1 || data[0].ProjectID == 0 {
		return 0, errors.Wrap(resultErr, url)
	}

	id := data[0].ProjectID
	ch <- prometheus.MustNewConstMetric(projectsRefInfo, prometheus.GaugeValue,
		1, projectsUrl, "GET")

	var result projectsJson

	url = projectsUrl + "/" + strconv.Itoa(id)
	body, err = client.request(url)
	if err != nil {
		return 0, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	if result.ProjectID == 0 {
		return 0, errors.Wrap(resultErr, url)
	}

	ch <- prometheus.MustNewConstMetric(projectsRefInfo, prometheus.GaugeValue,
		1, "/projects/{project_id}", "GET")

	return id, nil
}


func projectsLogs(id int,client *HarborClient, ch chan<- prometheus.Metric) error {
	var data []projectsJson
	url := fmt.Sprintf("/projects/%d/logs?page_size=1", id)
	body, err := client.request(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if len(data) != 1 || data[0].ProjectID == 0 {
		return errors.Wrap(resultErr, url)
	}

	ch <- prometheus.MustNewConstMetric(projectsRefInfo, prometheus.GaugeValue,
		1, "/projects/{project_id}/logs", "GET")

	return nil
}

func projectsMetadata(id int,client *HarborClient, ch chan<- prometheus.Metric) error {
	var data metadataJson
	url := fmt.Sprintf("/projects/%d/metadatas", id)
	body, err := client.request(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if len(data.Public) == 0 {
		return fmt.Errorf("cannot find the metadatas by %s", url)
	}

	ch <- prometheus.MustNewConstMetric(projectsRefInfo, prometheus.GaugeValue,
		1, "/projects/{project_id}/metadatas", "GET")

	url = fmt.Sprintf("/projects/%d/metadatas/public", id)
	body, err = client.request(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if len(data.Public) == 0 {
		return errors.Wrap(resultErr, url)
	}

	ch <- prometheus.MustNewConstMetric(projectsRefInfo, prometheus.GaugeValue,
		1, "/projects/{project_id}/metadatas/{meta_name}", "GET")

	return nil
}


func projectsMembers(id int, client *HarborClient, ch chan<- prometheus.Metric) error {
	var data []subInsJson
	url := fmt.Sprintf("/projects/%d/members", id)
	body, err := client.request(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if len(data) == 0 || data[0].Id == 0 { // will response the all members
		return errors.Wrap(resultErr, url)
	}

	ch <- prometheus.MustNewConstMetric(projectsRefInfo, prometheus.GaugeValue,
		1, "/projects/{project_id}/members", "GET")


	var result subInsJson
	// some version (e.g., v1.8.1 https://github.com/goharbor/harbor/issues/12273), It will return 403
	url = fmt.Sprintf("/projects/%d/members/%d", id, id)
	body, err = client.request(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Id == 0 {
		return errors.Wrap(resultErr, url)
	}

	ch <- prometheus.MustNewConstMetric(projectsRefInfo, prometheus.GaugeValue,
		1, "/projects/{project_id}/members/{mid}", "GET")


	return nil
}

type repoJson struct {
	subInsJson
	Name string `json:"name"`
}

// first arg must be project_id
func reposQuery(id int, client *HarborClient, ch chan<- prometheus.Metric) error {
	var data []repoJson
	url := "/repositories?project_id=" + strconv.Itoa(id)
	body, err := client.request(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if len(data) != 1 || len(data[0].Name) == 0 {
		return errors.Wrap(resultErr, url)
	}

	ch <- prometheus.MustNewConstMetric(reposRefInfo, prometheus.GaugeValue,
		1, "/repositories", "GET")

	return nil
}

// TODO
//  enable `notary_signer` for /repositories/{repo_name}/signatures
//  tags always return the all tags https://github.com/goharbor/harbor/issues/12279
//func tagsQuery(repo string, client *HarborClient, ch chan<- prometheus.Metric)  error {
//	var data []repoJson
//	url := "/repositories/" + repo + "/tags"
//	body, err := client.request(url)
//	if err != nil {
//		return err
//	}
//
//	if err := json.Unmarshal(body, &data); err != nil {
//		return err
//	}
//
//	if len(data) != 1 || len(data[0].Name) == 0 {
//		return fmt.Errorf("cannot find the repo by %s", url)
//	}
//
//	ch <- prometheus.MustNewConstMetric(reposRefInfo, prometheus.GaugeValue,
//		1, "/repositories/{repo_name}/tags", "GET")
//
//	return nil
//}




func reposTop(client *HarborClient, ch chan<- prometheus.Metric) error {
	var data []repoJson
	url := "/repositories/top" + "?count=1"
	body, err := client.request(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if len(data) != 1 || len(data[0].Name) == 0 {
		return fmt.Errorf("cannot find the repo by %s", url)
	}

	ch <- prometheus.MustNewConstMetric(reposRefInfo, prometheus.GaugeValue,
		1, "/repositories/top", "GET")

	return  nil
}











//const (
//	projectsUrl  = "/projects"
//	projectIDUrl = "/repositories?project_id="
//)
//
//var (
//	repositoriesPullCount = prometheus.NewDesc(
//		prometheus.BuildFQName(namespace, "", "repositories_pull_total"),
//		"Get public repositories which are accessed most.).",
//		[]string{"repo_name", "public"}, nil,
//	)
//	repositoriesStarCount = prometheus.NewDesc(
//		prometheus.BuildFQName(namespace, "", "repositories_star_total"),
//		"Get public repositories which are accessed most.).",
//		[]string{"repo_name", "public"}, nil,
//	)
//	repositoriesTagCount = prometheus.NewDesc(
//		prometheus.BuildFQName(namespace, "", "repositories_tags_total"),
//		"Get public repositories which are accessed most.).",
//		[]string{"repo_name", "public"}, nil,
//	)
//)
//
//type ScrapeProjects struct{}
//
//// Name of the Scraper. Should be unique.
//func (ScrapeProjects) Name() string {
//	return "projects"
//}
//
//// Help describes the role of the Scraper.
//func (ScrapeProjects) Help() string {
//	return "Collect the projects"
//}
//
//// Scrape collects data from client and sends it over channel as prometheus metric.
//func (s ScrapeProjects) Scrape(client *HarborClient, ch chan<- prometheus.Metric) error {
//	var data []projectsJson
//	body, err := client.request(projectsUrl)
//	if err != nil {
//		return err
//	}
//
//	if err := json.Unmarshal(body, &data); err != nil {
//		return err
//	}
//
//	for i := 0; i < len(data); i++ {
//		var repoData []repositoriesJson
//
//		repoBody, err := client.request(projectIDUrl + strconv.Itoa(data[i].ProjectID))
//		if err != nil {
//			log.WithField("scraper", s.Name()).Error(err)
//			continue
//		}
//		if err := json.Unmarshal(repoBody, &repoData); err != nil {
//			log.WithField("scraper", s.Name()).Error(err)
//			continue
//		}
//		for j := 0; j < len(repoData); j++ {
//			ch <- prometheus.MustNewConstMetric(repositoriesPullCount, prometheus.GaugeValue,
//				repoData[j].PullCount, repoData[j].Name, data[i].Metadata.Public)
//			ch <- prometheus.MustNewConstMetric(repositoriesStarCount, prometheus.GaugeValue,
//				repoData[j].StarCount, repoData[j].Name, data[i].Metadata.Public)
//			ch <- prometheus.MustNewConstMetric(repositoriesTagCount, prometheus.GaugeValue,
//				repoData[j].TagsCount, repoData[j].Name, data[i].Metadata.Public)
//		}
//
//	}
//
//	return nil
//}

//type projectsJson struct {
//	ProjectID int `json:"project_id"`
	//OwnerID           int       `json:"owner_id"`
	//Name              string    `json:"name"`
	//CreationTime      time.Time `json:"creation_time"`
	//UpdateTime        time.Time `json:"update_time"`
	//Deleted           int       `json:"deleted"`
	//OwnerName         string    `json:"owner_name"`
	//Togglable         bool      `json:"togglable"`
	//CurrentUserRoleID int       `json:"current_user_role_id"`
	//RepoCount         int       `json:"repo_count"`
//	Metadata struct {
//		Public string `json:"public"`
//	} `json:"metadata"`
//}

//type repositoriesJson struct {
	//ID           int           `json:"id"`
	//Name string `json:"name"`
	//ProjectID    int           `json:"project_id"`
	//Description  string        `json:"description"`
	//PullCount float64 `json:"pull_count"`
	//StarCount float64 `json:"star_count"`
	//TagsCount float64 `json:"tags_count"`
	//Labels       []interface{} `json:"labels"`
	//CreationTime time.Time     `json:"creation_time"`
	//UpdateTime   time.Time     `json:"update_time"`
//}