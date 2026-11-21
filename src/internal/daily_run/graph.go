package daily_run

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"time"
)

type NodeType string

const (
	NodeCombat   NodeType = "combat"
	NodeElite    NodeType = "elite"
	NodeMiniBoss NodeType = "mini_boss"
	NodeChest    NodeType = "chest"
	NodeRest     NodeType = "rest"
	NodeBoss     NodeType = "boss"
)

type GraphNode struct {
	ID          string   `json:"id"`
	Floor       uint8    `json:"floor"`
	Type        NodeType `json:"type"`
	NextNodeIDs []string `json:"next_node_ids"`
	// For chest/event rooms, DC is the d20 difficulty class and CheckStat is the stat used.
	CheckStat string `json:"check_stat,omitempty"`
	CheckDC   int    `json:"check_dc,omitempty"`
}

type DungeonGraph struct {
	RunDate     string               `json:"run_date"`
	Tier        uint8                `json:"tier"`
	TotalFloors uint8                `json:"total_floors"`
	StartNodeID string               `json:"start_node_id"`
	Nodes       []GraphNode          `json:"nodes"`
	nodeIndex   map[string]GraphNode `json:"-"`
}

func (g *DungeonGraph) NodeByID(id string) (GraphNode, bool) {
	if g.nodeIndex == nil {
		g.nodeIndex = make(map[string]GraphNode, len(g.Nodes))
		for _, n := range g.Nodes {
			g.nodeIndex[n.ID] = n
		}
	}
	n, ok := g.nodeIndex[id]
	return n, ok
}

func TotalFloorsForTier(tier uint8) uint8 {
	if tier < 1 || tier > 5 {
		return 10
	}
	return floorCountPerTier[tier-1]
}

func miniBossFloorsForTier(tier uint8) map[uint8]bool {
	switch tier {
	case 1:
		return map[uint8]bool{5: true}
	case 2:
		return map[uint8]bool{8: true}
	case 3:
		return map[uint8]bool{7: true, 14: true}
	case 4:
		return map[uint8]bool{8: true, 17: true}
	case 5:
		return map[uint8]bool{10: true, 20: true}
	default:
		return map[uint8]bool{}
	}
}

var eventCheckStats = []string{"dexterity", "intelligence", "wisdom", "strength", "charisma"}

// GenerateDungeon deterministically creates the DAG for a given UTC date and tier (1..5).
// All players running that tier on that date receive the exact same dungeon layout.
func GenerateDungeon(runDate time.Time, tier uint8) *DungeonGraph {
	dateStr := runDate.UTC().Format("2006-01-02")
	totalFloors := TotalFloorsForTier(tier)
	miniBossFloors := miniBossFloorsForTier(tier)

	sum := sha256.Sum256([]byte(fmt.Sprintf("daily-dungeon:%s:tier:%d", dateStr, tier)))
	seed1 := binary.LittleEndian.Uint64(sum[0:8])
	seed2 := binary.LittleEndian.Uint64(sum[8:16])
	rng := rand.New(rand.NewPCG(seed1, seed2))

	floorNodes := make([][]GraphNode, totalFloors)

	for f := uint8(1); f <= totalFloors; f++ {
		idx := f - 1
		switch {
		case f == 1:
			floorNodes[idx] = []GraphNode{{
				ID:    "f1_n1",
				Floor: 1,
				Type:  NodeCombat,
			}}
		case f == totalFloors:
			floorNodes[idx] = []GraphNode{{
				ID:    fmt.Sprintf("f%d_boss", f),
				Floor: f,
				Type:  NodeBoss,
			}}
		case miniBossFloors[f]:
			floorNodes[idx] = []GraphNode{{
				ID:    fmt.Sprintf("f%d_miniboss", f),
				Floor: f,
				Type:  NodeMiniBoss,
			}}
		case f == totalFloors-1 || miniBossFloors[f+1]:
			// Guarantee a campfire rest room option right before a mini-boss or final boss.
			nodeA := GraphNode{
				ID:    fmt.Sprintf("f%d_n1", f),
				Floor: f,
				Type:  NodeRest,
			}
			nodeB := GraphNode{
				ID:    fmt.Sprintf("f%d_n2", f),
				Floor: f,
				Type:  NodeElite,
			}
			floorNodes[idx] = []GraphNode{nodeA, nodeB}
		default:
			branchCount := pickBranchCount(rng)
			nodes := make([]GraphNode, branchCount)
			for b := 0; b < branchCount; b++ {
				nType := pickRoomType(rng, f)
				node := GraphNode{
					ID:    fmt.Sprintf("f%d_n%d", f, b+1),
					Floor: f,
					Type:  nType,
				}
				if nType == NodeChest {
					node.CheckStat = eventCheckStats[rng.IntN(len(eventCheckStats))]
					node.CheckDC = 10 + int(tier)*2 + rng.IntN(4)
				}
				nodes[b] = node
			}
			floorNodes[idx] = nodes
		}
	}

	// Connect each floor to the next floor so every node has >= 1 outgoing edge
	// and every node on the next floor has >= 1 incoming edge.
	for f := 0; f < int(totalFloors)-1; f++ {
		curr := floorNodes[f]
		next := floorNodes[f+1]
		incoming := make([]bool, len(next))

		for i := range curr {
			// Primary edge maps proportionally to next floor.
			primary := (i * len(next)) / len(curr)
			curr[i].NextNodeIDs = append(curr[i].NextNodeIDs, next[primary].ID)
			incoming[primary] = true

			// Optional second branching edge for tactical choice.
			if len(next) > 1 {
				alt := (primary + 1) % len(next)
				if len(curr) == 1 || rng.IntN(2) == 0 {
					curr[i].NextNodeIDs = append(curr[i].NextNodeIDs, next[alt].ID)
					incoming[alt] = true
				}
			}
		}

		// Ensure no unreachable orphan nodes on next floor.
		for j, hasIn := range incoming {
			if !hasIn {
				parentIdx := j % len(curr)
				curr[parentIdx].NextNodeIDs = append(curr[parentIdx].NextNodeIDs, next[j].ID)
			}
		}

		floorNodes[f] = curr
	}

	allNodes := make([]GraphNode, 0, int(totalFloors)*2)
	index := make(map[string]GraphNode)
	for _, fn := range floorNodes {
		for _, n := range fn {
			allNodes = append(allNodes, n)
			index[n.ID] = n
		}
	}

	return &DungeonGraph{
		RunDate:     dateStr,
		Tier:        tier,
		TotalFloors: totalFloors,
		StartNodeID: "f1_n1",
		Nodes:       allNodes,
		nodeIndex:   index,
	}
}

func pickBranchCount(rng *rand.Rand) int {
	roll := rng.IntN(100)
	switch {
	case roll < 30:
		return 1
	case roll < 80:
		return 2
	default:
		return 3
	}
}

func pickRoomType(rng *rand.Rand, floor uint8) NodeType {
	// Periodically offer a rest room on longer dungeons.
	if floor%6 == 0 {
		if rng.IntN(2) == 0 {
			return NodeRest
		}
	}
	roll := rng.IntN(100)
	switch {
	case roll < 45:
		return NodeCombat
	case roll < 65:
		return NodeElite
	case roll < 85:
		return NodeChest
	default:
		return NodeRest
	}
}

// DeterministicEnemyIndex returns a stable index into a list of enemies for a
// specific daily dungeon room so everyone fights the same monster in that room.
func DeterministicEnemyIndex(runDate time.Time, tier uint8, nodeID string, poolLen int) int {
	if poolLen <= 1 {
		return 0
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("enemy:%s:%d:%s", runDate.UTC().Format("2006-01-02"), tier, nodeID)))
	v := binary.LittleEndian.Uint64(sum[0:8])
	return int(v % uint64(poolLen))
}
