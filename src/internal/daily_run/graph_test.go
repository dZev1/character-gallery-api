package daily_run

import (
	"testing"
	"time"
)

func TestGenerateDungeon_DeterministicAndFloorCounts(t *testing.T) {
	date := time.Date(2026, 11, 20, 15, 30, 0, 0, time.UTC)

	expectedFloors := map[uint8]uint8{
		1: 10,
		2: 15,
		3: 20,
		4: 25,
		5: 30,
	}

	for tier, wantFloors := range expectedFloors {
		g1 := GenerateDungeon(date, tier)
		g2 := GenerateDungeon(date, tier)

		if g1.TotalFloors != wantFloors {
			t.Fatalf("tier %d: got total floors %d, want %d", tier, g1.TotalFloors, wantFloors)
		}
		if len(g1.Nodes) != len(g2.Nodes) {
			t.Fatalf("tier %d: non-deterministic node count (%d vs %d)", tier, len(g1.Nodes), len(g2.Nodes))
		}
		for i := range g1.Nodes {
			if g1.Nodes[i].ID != g2.Nodes[i].ID || g1.Nodes[i].Type != g2.Nodes[i].Type {
				t.Fatalf("tier %d node %d mismatch: %+v vs %+v", tier, i, g1.Nodes[i], g2.Nodes[i])
			}
		}

		// Verify every node except the final boss has at least one outgoing edge
		// and all edges point to valid nodes on floor + 1.
		for _, n := range g1.Nodes {
			if n.Floor == wantFloors {
				if n.Type != NodeBoss {
					t.Errorf("tier %d final floor node is %s, want boss", tier, n.Type)
				}
				if len(n.NextNodeIDs) != 0 {
					t.Errorf("tier %d boss node should have no outgoing edges", tier)
				}
				continue
			}
			if len(n.NextNodeIDs) == 0 {
				t.Errorf("tier %d node %s (floor %d) has no outgoing edges", tier, n.ID, n.Floor)
			}
			for _, nextID := range n.NextNodeIDs {
				nextNode, ok := g1.NodeByID(nextID)
				if !ok {
					t.Errorf("tier %d node %s points to missing node %s", tier, n.ID, nextID)
				}
				if nextNode.Floor != n.Floor+1 {
					t.Errorf("tier %d node %s (floor %d) points to %s (floor %d)", tier, n.ID, n.Floor, nextID, nextNode.Floor)
				}
			}
		}
	}
}
