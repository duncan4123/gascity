package beads_test

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/gastownhall/gascity/internal/beads"
	"github.com/gastownhall/gascity/internal/beads/beadstest"
)

const backendHelperEnv = "GC_GASCITY_BACKEND_CONFORMANCE_HELPER"

func TestGascityBackendStoreConformance(t *testing.T) {
	if os.Getenv(backendHelperEnv) == "1" {
		runBackendHelper()
		return
	}
	var stores []*beads.GascityBackendStore
	t.Cleanup(func() {
		for _, store := range stores {
			_ = store.CloseStore()
		}
	})
	beadstest.RunStoreTests(t, func() beads.Store {
		store, err := beads.OpenGascityBackendStore(context.Background(), beads.GascityBackendStoreConfig{
			Command: os.Args[0],
			Args:    []string{"-test.run=^TestGascityBackendStoreConformance$"},
			Env:     map[string]string{backendHelperEnv: "1"},
		})
		if err != nil {
			panic(err)
		}
		stores = append(stores, store)
		return store
	})
}

type helperRequest struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type helperResponse struct {
	ID     string `json:"id,omitempty"`
	OK     bool   `json:"ok"`
	Result any    `json:"result,omitempty"`
	Error  any    `json:"error,omitempty"`
}

func runBackendHelper() {
	store := beads.NewMemStore()
	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(helperResponse{OK: true, Result: map[string]any{"protocol": beads.GascityBackendProtocolV1Alpha1}})
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var req helperRequest
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			_ = enc.Encode(helperErr(req.ID, "bad_request", err))
			continue
		}
		resp := handleHelperRequest(store, req)
		_ = enc.Encode(resp)
	}
}

func helperOK(id string, result any) helperResponse {
	return helperResponse{ID: id, OK: true, Result: result}
}

func helperErr(id, code string, err error) helperResponse {
	return helperResponse{ID: id, OK: false, Error: map[string]string{"code": code, "message": err.Error()}}
}

func decodeHelper(raw json.RawMessage, out any) error { return json.Unmarshal(raw, out) }

func handleHelperRequest(store beads.Store, req helperRequest) helperResponse {
	var p struct {
		ID           string          `json:"id"`
		IssueID      string          `json:"issue_id"`
		DependsOnID  string          `json:"depends_on_id"`
		Issue        json.RawMessage `json:"issue"`
		Updates      map[string]any  `json:"updates"`
		AddLabels    []string        `json:"add_labels"`
		RemoveLabels []string        `json:"remove_labels"`
		ParentID     *string         `json:"parent_id"`
		Key          string          `json:"key"`
		Value        string          `json:"value"`
		Dependency   beads.Dep       `json:"dependency"`
		Filter       map[string]any  `json:"filter"`
	}
	if err := decodeHelper(req.Params, &p); err != nil {
		return helperErr(req.ID, "bad_request", err)
	}
	switch req.Method {
	case "open":
		return helperOK(req.ID, map[string]string{"session_id": "test"})
	case "close", "ping":
		return helperOK(req.ID, map[string]bool{"ok": true})
	case "create_issue":
		var wire struct {
			beads.Bead
			Sender string `json:"sender"`
		}
		if err := json.Unmarshal(p.Issue, &wire); err != nil {
			return helperErr(req.ID, "bad_request", err)
		}
		b := wire.Bead
		b.From = wire.Sender
		if b.ParentID == "" {
			for _, dep := range b.Dependencies {
				if dep.Type == "parent-child" {
					b.ParentID = dep.DependsOnID
					break
				}
			}
		}
		created, err := store.Create(b)
		if err != nil {
			return helperErr(req.ID, "storage_error", err)
		}
		return helperOK(req.ID, created)
	case "get_issue":
		b, err := store.Get(p.ID)
		if err != nil {
			return helperStoreErr(req.ID, err)
		}
		return helperOK(req.ID, b)
	case "update_issue":
		opts := updateOptsFromHelper(p.Updates)
		opts.Labels, opts.RemoveLabels, opts.ParentID = p.AddLabels, p.RemoveLabels, p.ParentID
		if err := store.Update(p.ID, opts); err != nil {
			return helperStoreErr(req.ID, err)
		}
		b, _ := store.Get(p.ID)
		return helperOK(req.ID, b)
	case "close_issue":
		if err := store.Close(p.ID); err != nil {
			return helperStoreErr(req.ID, err)
		}
		return helperOK(req.ID, map[string]bool{"closed": true})
	case "reopen_issue":
		if err := store.Reopen(p.ID); err != nil {
			return helperStoreErr(req.ID, err)
		}
		return helperOK(req.ID, map[string]bool{"reopened": true})
	case "delete_issue":
		if err := store.Delete(p.ID); err != nil {
			return helperStoreErr(req.ID, err)
		}
		return helperOK(req.ID, map[string]bool{"deleted": true})
	case "search_issues":
		all, err := store.List(beads.ListQuery{AllowScan: true, IncludeClosed: true, TierMode: beads.TierIssues})
		if err != nil {
			return helperStoreErr(req.ID, err)
		}
		return helperOK(req.ID, all)
	case "list_wisps":
		all, err := store.List(beads.ListQuery{AllowScan: true, IncludeClosed: true, TierMode: beads.TierBoth})
		if err != nil {
			return helperStoreErr(req.ID, err)
		}
		wisps := all[:0]
		for _, b := range all {
			if b.Ephemeral {
				wisps = append(wisps, b)
			}
		}
		return helperOK(req.ID, wisps)
	case "ready_work":
		all, err := store.Ready(beads.ReadyQuery{TierMode: beads.TierBoth})
		if err != nil {
			return helperStoreErr(req.ID, err)
		}
		return helperOK(req.ID, all)
	case "slot_set":
		if err := store.SetMetadata(p.IssueID, p.Key, p.Value); err != nil {
			return helperStoreErr(req.ID, err)
		}
		return helperOK(req.ID, map[string]bool{"set": true})
	case "add_dependency":
		if err := store.DepAdd(p.Dependency.IssueID, p.Dependency.DependsOnID, p.Dependency.Type); err != nil {
			return helperStoreErr(req.ID, err)
		}
		return helperOK(req.ID, map[string]bool{"added": true})
	case "remove_dependency":
		if err := store.DepRemove(p.IssueID, p.DependsOnID); err != nil {
			return helperStoreErr(req.ID, err)
		}
		return helperOK(req.ID, map[string]bool{"removed": true})
	case "get_dependencies":
		deps, err := store.DepList(p.IssueID, "down")
		if err != nil {
			return helperStoreErr(req.ID, err)
		}
		return helperOK(req.ID, deps)
	case "get_dependents":
		deps, err := store.DepList(p.IssueID, "up")
		if err != nil {
			return helperStoreErr(req.ID, err)
		}
		return helperOK(req.ID, deps)
	default:
		return helperErr(req.ID, "unknown_method", fmt.Errorf("unknown method %s", req.Method))
	}
}

func helperStoreErr(id string, err error) helperResponse {
	code := "storage_error"
	if errors.Is(err, beads.ErrNotFound) {
		code = "not_found"
	}
	return helperErr(id, code, err)
}

func updateOptsFromHelper(fields map[string]any) beads.UpdateOpts {
	var out beads.UpdateOpts
	if v, ok := fields["title"].(string); ok {
		out.Title = &v
	}
	if v, ok := fields["status"].(string); ok {
		out.Status = &v
	}
	if v, ok := fields["issue_type"].(string); ok {
		out.Type = &v
	}
	if v, ok := fields["description"].(string); ok {
		out.Description = &v
	}
	if v, ok := fields["assignee"].(string); ok {
		out.Assignee = &v
	}
	if v, ok := fields["priority"].(float64); ok {
		n := int(v)
		out.Priority = &n
	}
	if v, ok := fields["metadata"].(map[string]any); ok {
		out.Metadata = make(map[string]string, len(v))
		for key, raw := range v {
			if value, ok := raw.(string); ok {
				out.Metadata[key] = value
			}
		}
	}
	return out
}
