package awssync

import "sshh/internal/model"

// SyncAction describes what a SyncItem would do if applied.
type SyncAction int

const (
	// SyncNew adds a server that has no matching name in the existing inventory.
	SyncNew SyncAction = iota
	// SyncUpdateIP updates the Host of an existing server whose AWS private IP has changed.
	SyncUpdateIP
	// SyncStale flags a previously AWS-managed server whose EC2 instance no
	// longer appears in the fetch (terminated, stopped, or untagged).
	SyncStale
)

// SyncItem is a single pending change produced by Diff.
type SyncItem struct {
	Action        SyncAction
	Server        model.Server // record to write (or remove, for SyncStale) if this item is applied
	OldHost       string       // previous Host; only set for SyncUpdateIP
	ExistingIndex int          // index into the existing slice; meaningful for SyncUpdateIP and SyncStale
}

// Diff compares fetched AWS instances against the existing inventory, matching
// by Server.Name (case-sensitive, same semantics as Config.FindByName).
//
// Instances with no matching name become SyncNew items using defaultUser.
// Instances matching an existing server with a different Host become
// SyncUpdateIP items that preserve the existing User/Port/Key/Tags. Matches
// with an unchanged Host are omitted. Both SyncNew and SyncUpdateIP results
// are marked AWSManaged so a later sync can track them.
//
// Existing servers that are AWSManaged (added or IP-updated by a previous
// sync) but whose name is absent from the current fetch become SyncStale
// items — their EC2 instance is presumed terminated, stopped, or untagged.
// Servers that were never synced from AWS are never flagged stale, even if
// their name happens not to match anything in the current fetch.
func Diff(existing []model.Server, fetched []model.Server, defaultUser string) []SyncItem {
	nameToIdx := make(map[string]int, len(existing))
	for i, s := range existing {
		nameToIdx[s.Name] = i
	}

	fetchedNames := make(map[string]bool, len(fetched))
	var items []SyncItem
	for _, f := range fetched {
		fetchedNames[f.Name] = true

		idx, ok := nameToIdx[f.Name]
		if !ok {
			items = append(items, SyncItem{
				Action: SyncNew,
				Server: model.Server{
					Name:       f.Name,
					Host:       f.Host,
					User:       defaultUser,
					Port:       22,
					AWSManaged: true,
				},
			})
			continue
		}

		cur := existing[idx]
		if cur.Host == f.Host {
			continue
		}
		updated := cur
		updated.Host = f.Host
		updated.AWSManaged = true
		items = append(items, SyncItem{
			Action:        SyncUpdateIP,
			Server:        updated,
			OldHost:       cur.Host,
			ExistingIndex: idx,
		})
	}

	for i, s := range existing {
		if !s.AWSManaged || fetchedNames[s.Name] {
			continue
		}
		items = append(items, SyncItem{
			Action:        SyncStale,
			Server:        s,
			ExistingIndex: i,
		})
	}

	return items
}
