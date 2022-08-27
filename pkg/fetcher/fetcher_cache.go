package fetcher

import (
	"encoding/json"
	"kubectlfzf/pkg/k8s/resources"
	"kubectlfzf/pkg/util"
	"net/http"
	"os"
	"path"
	"time"

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

func (f *Fetcher) getLastModifiedTimesPath() string {
	return path.Join(f.fetcherCachePath, f.GetContext(), "lastModified")
}

func (f *Fetcher) getCachedNamespace() string {
	filePath := path.Join(f.fetcherCachePath, f.GetContext(), "fzfNamespace")
	b, err := os.ReadFile(filePath)
	if err != nil {
		logrus.Warnf("Couldn't read namespace cache path %s", filePath)
		return ""
	}
	return string(b)
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

func (f *Fetcher) writeCachedNamespace(ns string) error {
	cacheDir, err := f.createCacheDir()
	if err != nil {
		return err
	}
	filePath := path.Join(cacheDir, "fzfNamespace")
	logrus.Infof("Writing kubectl fzf namespace '%s' in cache file %s", ns, filePath)
	err = os.WriteFile(filePath, []byte(ns), 0644)
	return err
}

func (f *Fetcher) loadLastModifiedTimes() (map[string]time.Time, error) {
	lastModifiedTimesPath := f.getLastModifiedTimesPath()
	if !util.FileExists(lastModifiedTimesPath) {
		return map[string]time.Time{}, nil
	}
	b, err := os.ReadFile(lastModifiedTimesPath)
	if err != nil {
		return nil, err
	}
	var lastModifiedTimes map[string]time.Time
	err = json.Unmarshal(b, &lastModifiedTimes)
	return lastModifiedTimes, err
}

func (f *Fetcher) updateLastModifiedTimes(r resources.ResourceType, newTime time.Time) error {
	logrus.Infof("Updating last modified times for %s", r)
	lastModifiedTimes, err := f.loadLastModifiedTimes()
	if err != nil {
		return err
	}
	lastModifiedTimes[r.String()] = newTime
	b, err := json.Marshal(lastModifiedTimes)
	return os.WriteFile(f.getLastModifiedTimesPath(), b, 0644)
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
	return f.updateLastModifiedTimes(r, lastModifiedTime)
}

func (f *Fetcher) getResourceFromCache(r resources.ResourceType) (map[string]resources.K8sResource, error) {
	cacheFile := path.Join(f.fetcherCachePath, f.GetContext(), r.String())
	resources := map[string]resources.K8sResource{}
	err := util.LoadGobFromFile(&resources, cacheFile)
	return resources, err
}

func (f *Fetcher) checkLocalCache(endpoint string, r resources.ResourceType) (map[string]resources.K8sResource, error) {
	cacheFile := path.Join(f.fetcherCachePath, f.GetContext(), r.String())
	finfo, err := os.Stat(cacheFile)
	resources := map[string]resources.K8sResource{}
	if err != nil {
		logrus.Infof("No cache file %s present", cacheFile)
		return nil, nil
	}

	// A cache file is present
	deltaMod := time.Now().Sub(finfo.ModTime())
	if deltaMod <= f.minimumCache {
		logrus.Infof("Cache file present and was modified %s ago, using it", deltaMod)
		err := util.LoadGobFromFile(&resources, cacheFile)
		return resources, err
	}

	modifiedTimes, err := f.loadLastModifiedTimes()
	if err != nil {
		logrus.Infof("Couldn't read modified times, aborting use of cache file: %s", err)
		return nil, err
	}

	localLastModified, ok := modifiedTimes[r.String()]
	resourcePath := f.getResourceHttpPath(endpoint, r)
	if ok {
		headers, err := util.HeadFromHttpServer(resourcePath)
		if err != nil {
			return nil, errors.Wrapf(err, "error on head of %s", resourcePath)
		}
		lastModifiedTime, err := getLastModifiedFromHeader(headers)
		// No change, load from cache file
		if lastModifiedTime == localLastModified {
			err = util.LoadGobFromFile(&resources, cacheFile)
			return resources, err
		}
		logrus.Infof("Resource %s was modified on server, pulling new version: old modified time %s, new modified time %s", r, localLastModified, lastModifiedTime)
	} else {
		logrus.Infof("No modified times for %s, pulling it from server", r)
	}
	return nil, err
}
