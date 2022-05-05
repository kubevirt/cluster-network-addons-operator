package release

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/Masterminds/semver"
)

const KnmstateReleasesUrl = "https://api.github.com/repos/nmstate/kubernetes-nmstate/releases"

type Release struct {
	TagName string `json:"tag_name"`
}

func newHttpClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 2,
	}
}

func newGetRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func GetUrlData(url string) ([]byte, error) {
	req, err := newGetRequest(url)
	if err != nil {
		return []byte{}, err
	}

	resp, err := newHttpClient().Do(req)
	if err != nil {
		return []byte{}, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

func ParseReleases(data []byte) ([]Release, error) {
	r := make([]Release, 0)
	jsonErr := json.Unmarshal(data, &r)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return r, nil
}

func SortReleasesBySemver(releases []Release) []*semver.Version {
	sv := make([]*semver.Version, len(releases))
	for i, r := range releases {
		v, err := semver.NewVersion(r.TagName)
		if err != nil {
			fmt.Printf("Error parsing version: %s\n", err)
		}
		sv[i] = v
	}

	sort.Sort(semver.Collection(sv))
	return sv
}
