package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"sync"
    "runtime"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
    screenWidth = 512
    screenHeight = 384
)

type Rule struct {
    neighborhood int
    min int
    max *int
    nextState bool
}

// Returns where a value is inside a Rule's interval
func (r *Rule) Contains(val int) bool {
    if r.max != nil {
        return val >= r.min && val <= *(r.max)
    } else {
        return val >= r.min
    }
}

type Coordinate struct {
    x, y int
}

type Neighborhood struct {
    validNeighbors int
    neighbors []Coordinate
}

type EvolutionRules struct {
    numNeighborhoods int
    neighborhoods []Neighborhood
    rulesList []Rule
}

type World struct {
    grid, nextGrid []bool
    width, height int
    rules EvolutionRules
}

func InitializeWorld(width, height int) *World {
    w := &World {
        grid: make([]bool, width*height),
        nextGrid: make([]bool, width*height),
        width: width,
        height: height,
        rules: readNeighborhoods(),
    }

    for i := range w.grid {
        w.grid[i] = rand.IntN(100) < 30
    }

    return w
}

func (w *World) Update() {
    var wg sync.WaitGroup
    chunkSize := w.height / runtime.NumCPU()

    for startY := 0; startY < w.height; startY += chunkSize {
        wg.Add(1)
        endY := startY + chunkSize
        if endY > w.height {
            endY = w.height
        }

        go func(startY, endY int) {
            defer wg.Done()
            sums := make([]int, len(w.rules.neighborhoods))


            for y := startY; y < endY; y++ {
                for x := 0; x < w.width; x++ {
                    // calculate % of alive neighbors
                    for i, neighborhood := range w.rules.neighborhoods {
                        nCount := neighborCount(w.grid, w.width, w.height, x, y, &neighborhood)
                        sums[i] = nCount
                    }

                    // calculate next state based on EvolutionRules
                    nextState := w.grid[y*w.width+x]
                    for _, rule := range w.rules.rulesList {
                        if rule.Contains(sums[rule.neighborhood]) {
                            nextState = rule.nextState
                        }
                    }

                    w.nextGrid[y*w.width+x] = nextState
                }
            }
        }(startY, endY)
    }

    wg.Wait()

    // swap grids
    w.grid, w.nextGrid = w.nextGrid, w.grid
}

func neighborCount(a []bool, width, height, x, y int, n *Neighborhood) int {
    c := 0
    for _, coord := range n.neighbors {
        newX := x + coord.x
        newY := y + coord.y

        if newX < 0 {
            newX += width
        } else if newX >= width {
            newX -= width
        }

        if newY < 0 {
            newY += height
        } else if newY >= height{
            newY -= height
        }
        if a[newY*width+newX] {
            c++
        }
    }
    return c
}

func (w *World) Draw(pix []byte) {
    for i, v := range w.grid {
        if v {
            pix[4*i], pix[4*i+1], pix[4*i+2], pix[4*i+3] = 0xff, 0xff, 0xff, 0xff
        } else {
            pix[4*i], pix[4*i+1], pix[4*i+2], pix[4*i+3] = 0, 0, 0, 0
        }
    }
}

type Game struct {
    world *World
    pixels []byte
}

func (g *Game) Update() error {
    g.world.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    if g.pixels == nil {
        g.pixels = make([]byte, screenWidth*screenHeight*4)
    }
    g.world.Draw(g.pixels)
    screen.WritePixels(g.pixels)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

type NeighborhoodMap [31][31] bool

// ParseBool converts "0" or "1" to false or true respectively.
func ParseBool(value string) (bool, error) {
	if value == "0" {
		return false, nil
	} else if value == "1" {
		return true, nil
	}
	return false, fmt.Errorf("invalid value: %s", value)
}

func readNeighborhoods() (EvolutionRules) {
<<<<<<< HEAD
=======
<<<<<<< HEAD
<<<<<<< HEAD
    rulesFilePath := "rules/example.txt"
    if len(os.Args) == 2 {
        rulesFilePath = os.Args[1]
    }

    file, err := os.Open(rulesFilePath)
=======
=======
>>>>>>> main
>>>>>>> dev_rulesFilesMod
    rulesFile := "rules/example.txt"
    if len(os.Args) == 2 {
        rulesFile = os.Args[1]
    }
    file, err := os.Open(rulesFile)
>>>>>>> 15b4f5d (changes)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    reader := bufio.NewReader(file)

    neighborhoods := make(map[int]NeighborhoodMap)
    var rules []Rule
    var currentNeighborhood int
    var currentRow int
    inRuleSection := false

    for {
        line, err  := reader.ReadString('\n')
        if err != nil {
            if err.Error() != "EOF" {
                fmt.Println("Error reading file: ", err)
            }
            break
        }

        line = strings.TrimSpace(line)

        if line == "[Rule]" {
            inRuleSection = true
        }

        if inRuleSection {
            if strings.Contains(line, "=") {
                parts := strings.Split(line, "=")
                ruleID := parts[0]
                ruleValues := strings.Fields(parts[1])

                if len(ruleValues) == 3 {
                    low, _ := strconv.Atoi(ruleValues[0])
                    next, _ := ParseBool(ruleValues[2])

                    if len(ruleID) >= 3 && strings.HasPrefix(ruleID, "S") {
                        neighborhoodID, _ := strconv.Atoi(string(ruleID[1]))
                        var high *int
                        if ruleValues[1] != "0" {
                            highVal, _ := strconv.Atoi(ruleValues[1])
                            high = &highVal
                        }


                        rules = append(rules, Rule {
                            neighborhood: neighborhoodID,
                            min: low,
                            max: high,
                            nextState: next,
                        })
                    } else {
                        fmt.Println("ERROR invalid ruleID format: ", ruleID)
                    }
                } else if len(ruleValues) == 2 {
                    low, _ := strconv.Atoi(ruleValues[0])
                    next, _ := ParseBool(ruleValues[1])

                    if len(ruleID) >= 3 && strings.HasPrefix(ruleID, "S") {
                        neighborhoodID, _ := strconv.Atoi(string(ruleID[1]))

                        rules = append(rules, Rule {
                            neighborhood: neighborhoodID,
                            min: low,
                            max: nil,
                            nextState: next,
                        })
                    } else {
                        fmt.Println("ERROR invalid ruleID format: ", ruleID)
                    }

                }
            }
        } else if strings.HasPrefix(line, "[N") {
            parts := strings.Split(line[1:len(line)-1], " ")
			neighborhoodID, _ := strconv.Atoi(strings.TrimPrefix(parts[0], "N"))
			rowID, _ := strconv.Atoi(parts[1])

			// Update the current neighborhood and row
			currentNeighborhood = neighborhoodID
			currentRow = rowID

			// If the neighborhood doesn't exist, initialize it
			if _, exists := neighborhoods[currentNeighborhood]; !exists {
				neighborhoods[currentNeighborhood] = NeighborhoodMap{}
			}
        } else if strings.HasPrefix(line, "N"){
            // Parse the line containing column values like N2 1=0
			parts := strings.Split(line, "=")
			column, _ := strconv.Atoi(strings.Fields(parts[0])[1])
			value := parts[1]


			// Convert the value to a boolean
			boolValue, err := ParseBool(value)
			if err != nil {
				fmt.Println("Error parsing boolean value:", err)
				continue
			}

			// Update the corresponding row and column in the current neighborhood
			neigh := neighborhoods[currentNeighborhood]
			neigh[currentRow][column] = boolValue
			neighborhoods[currentNeighborhood] = neigh
        }
    }

    relativeNeighborhoods := make([]Neighborhood, len(neighborhoods))
    for nID, neighborhood := range neighborhoods {
        var coords []Coordinate
        for rID, row := range neighborhood {
            for cID, col := range row {
                if col {
                    newX := cID - 15
                    newY := rID - 15
                    if newX == 0 && newY == 0 {
                        continue
                    }
                    coords = append(coords, Coordinate {
                        x: newX,
                        y: newY,
                    })
                }
            }
        }

        currentNeighborhood := Neighborhood{
            validNeighbors: len(coords),
            neighbors: coords,
        }
        relativeNeighborhoods[nID-1] = currentNeighborhood
    }

    return EvolutionRules{
        numNeighborhoods: len(relativeNeighborhoods),
        neighborhoods: relativeNeighborhoods,
        rulesList: rules,
    }

}

func main() {
    g := &Game {
        world: InitializeWorld(screenWidth, screenHeight),
    }

	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("MNCA Simulator")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
