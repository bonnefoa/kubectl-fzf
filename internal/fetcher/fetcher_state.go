package fetcher

import (
	"encoding/json"
	"os"
	"path"
	"time"

	"github.com/bonnefoa/kubectl-fzf/v3/internal/k8s/resources"
	"github.com/bonnefoa/kubectl-fzf/v3/internal/util"

	"github.com/sirupsen/logrus"
)

type FetcherState struct {
	statePath     string
	ContextStates map[string]*fetcherContextState
	hasChanged    bool
}

type fetcherContextState struct {
	FzfNamespace      string
	LastModifiedTimes map[resources.ResourceType]time.Time // Keep track of last modified times of a resource pulled
}

func newFetcherState(cachePath string) *FetcherState {
	return &FetcherState{
		statePath:     path.Join(cachePath, "fetcher_state"),
		ContextStates: map[string]*fetcherContextState{},
	}
}

func (f *FetcherState) getContextState(context string) *fetcherContextState {
	contextState, ok := f.ContextStates[context]
	if !ok {
		contextState = &fetcherContextState{
			LastModifiedTimes: map[resources.ResourceType]time.Time{},
			FzfNamespace:      "",
		}
		f.ContextStates[context] = contextState
	}
	return contextState
}

func (f *FetcherState) loadStateFromDisk() error {
	if !util.FileExists(f.statePath) {
		return nil
	}
	b, err := os.ReadFile(f.statePath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &f)
	return err
}

func (f *FetcherState) writeToDisk() error {
	if !f.hasChanged {
		return nil
	}
	b, err := json.Marshal(f)
	if err != nil {
		logrus.Errorf("Error while marshalling json; %s", err)
		return err
	}
	return os.WriteFile(f.statePath, b, 0644)
}

func (f *FetcherState) getLastModifiedTime(context string, r resources.ResourceType) *time.Time {
	contextState := f.getContextState(context)
	lastModified, ok := contextState.LastModifiedTimes[r]
	if !ok {
		return nil
	}
	return &lastModified
}

func (f *FetcherState) getFzfNamespace(context string) string {
	contextState := f.getContextState(context)
	return contextState.FzfNamespace
}

func (f *FetcherState) updateLastModifiedTimes(context string, r resources.ResourceType, newTime time.Time) {
	logrus.Infof("Updating last modified times for resource %s, state file %s", r, f.statePath)
	contextState := f.getContextState(context)
	contextState.LastModifiedTimes[r] = newTime
	f.hasChanged = true
}

func (f *FetcherState) updateNamespace(context string, namespace string) {
	contextState := f.getContextState(context)
	if contextState.FzfNamespace == namespace {
		return
	}
	logrus.Infof("Updating Namespace for %s to %s", context, namespace)
	f.hasChanged = true
	contextState.FzfNamespace = namespace
}
