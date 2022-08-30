package fetcher

import (
	"net/http"
	"os"
	"path"
	"time"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

func getLastModifiedFromHeader(headers http.Header) (time.Time, error) {
	lastModifiedStr := headers.Get("Last-Modified")
	lastModifiedTime, err := time.Parse(TimeFormat, lastModifiedStr)
	if err != nil {
		return lastModifiedTime, errors.Wrap(err, "invalid lastModified timestamp")
	}
	return lastModifiedTime, nil
}

func (f *Fetcher) createCacheDir() (string, error) {
	cacheDir := path.Join(f.fetcherCachePath, f.GetContext())
	logrus.Infof("Creating cache dir %s", cacheDir)
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return cacheDir, errors.Wrap(err, "error mkdirall")
	}
	return cacheDir, nil
}

func (f *Fetcher) writeResourceToCache(headers http.Header, b []byte, r resources.ResourceType) error {
	cacheDir, err := f.createCacheDir()
	if err != nil {
		return err
	}
	resourcePath := path.Join(cacheDir, r.String())
	logrus.Debugf("Caching resource in %s", resourcePath)
	err = os.WriteFile(resourcePath, b, 0644)
	if err != nil {
		return errors.Wrap(err, "error writing cache file")
	}
	lastModifiedTime, err := getLastModifiedFromHeader(headers)
	if err != nil {
		return err
	}
	f.fetcherState.updateLastModifiedTimes(f.GetContext(), r, lastModifiedTime)
	return nil
}

func (f *Fetcher) getResourceFromCache(r resources.ResourceType) (map[string]resources.K8sResource, error) {
	cacheFile := path.Join(f.fetcherCachePath, f.GetContext(), r.String())
	resources := map[string]resources.K8sResource{}
	err := util.LoadGobFromFile(&resources, cacheFile)
	return resources, err
}

func (f *Fetcher) checkRecentCache(r resources.ResourceType) (map[string]resources.K8sResource, error) {
	cacheFile := path.Join(f.fetcherCachePath, f.GetContext(), r.String())
	finfo, err := os.Stat(cacheFile)
	if err != nil {
		logrus.Infof("No cache file %s present", cacheFile)
		return nil, nil
	}

	// A cache file is present
	deltaMod := time.Now().Sub(finfo.ModTime())
	if deltaMod <= f.minimumCache {
		logrus.Infof("Cache file present and was modified %s ago, using it", deltaMod)
		resources := map[string]resources.K8sResource{}
		err := util.LoadGobFromFile(&resources, cacheFile)
		return resources, err
	}
	return nil, nil
}

func (f *Fetcher) checkHttpCache(endpoint string, r resources.ResourceType) (map[string]resources.K8sResource, error) {
	cacheFile := path.Join(f.fetcherCachePath, f.GetContext(), r.String())
	finfo, err := os.Stat(cacheFile)
	if err != nil {
		logrus.Infof("No cache file %s present", cacheFile)
		return nil, nil
	}

	// A cache file is present
	deltaMod := time.Now().Sub(finfo.ModTime())
	resources := map[string]resources.K8sResource{}
	if deltaMod <= f.minimumCache {
		logrus.Infof("Cache file present and was modified %s ago, using it", deltaMod)
		err := util.LoadGobFromFile(&resources, cacheFile)
		return resources, err
	}

	localLastModified := f.fetcherState.getLastModifiedTime(f.GetContext(), r)
	if localLastModified != nil {
		resourcePath := f.getResourceHttpPath(endpoint, r)
		headers, err := util.HeadFromHttpServer(resourcePath)
		if err != nil {
			return nil, errors.Wrapf(err, "error on head of %s", resourcePath)
		}
		lastModifiedTime, err := getLastModifiedFromHeader(headers)
		// No change, load from cache file
		if lastModifiedTime == *localLastModified {
			logrus.Infof("Cache has the same modified time %s, pulling %s data from local files", localLastModified, r)
			err = util.LoadGobFromFile(&resources, cacheFile)
			return resources, err
		}
		logrus.Infof("Resource %s was modified on server, pulling new version: old modified time %s, new modified time %s", r, localLastModified, lastModifiedTime)
	} else {
		logrus.Infof("No modified times for %s, pulling it from server", r)
	}
	return nil, err
}
