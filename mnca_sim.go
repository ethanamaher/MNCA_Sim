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
    "time"
    "runtime"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
    screenWidth = 512
    screenHeight = 384

    // Neighborhoods must be input as 31x31 for this to work
    // Rules files could be modified to give sizes before input so you can convert to relative coordinates correctly
    dx = 15
    dy = 15
)

type Cell struct {
    age int
    alive bool
}

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
    neighborCount int
    neighbors []Coordinate
}

type EvolutionRules struct {
    numNeighborhoods int
    neighborhoods []Neighborhood
    rulesList []Rule
}

type World struct {
    grid, nextGrid []Cell
    width, height int
    rules EvolutionRules
}

func InitializeWorld(width, height int) *World {
    w := &World {
        grid: make([]Cell, width*height),
        nextGrid: make([]Cell, width*height),
        width: width,
        height: height,
        rules: readNeighborhoods(),
    }

    for i := range w.grid {
        startsAlive := rand.IntN(100) < 30
        w.grid[i] = Cell {
            age : 1,
            alive: startsAlive,
        }
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

                    idx := y*w.width+x

                    // calculate next state based on EvolutionRules
                    nextState := w.grid[idx].alive
                    for _, rule := range w.rules.rulesList {
                        if rule.Contains(sums[rule.neighborhood]) {
                            nextState = rule.nextState
                        }
                    }

                    w.nextGrid[idx].alive = nextState

                    if nextState {
                        if w.nextGrid[idx].alive {
                            w.nextGrid[idx].age++
                        } else {
                            w.nextGrid[idx].age = 1
                        }
                    } else {
                        w.nextGrid[idx].age = 0
                    }
                }
            }
        }(startY, endY)
    }

    wg.Wait()

    // swap grids
    w.grid, w.nextGrid = w.nextGrid, w.grid
}

func neighborCount(a []Cell, width, height, x, y int, n *Neighborhood) int {
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
        if a[newY*width+newX].alive {
            c++
        }
    }
    return c
}

func (w *World) Draw(pix []byte) {
    for i, cell := range w.nextGrid {
        if cell.alive {
            switch cell.age {
            case 1:
                pix[4*i], pix[4*i+1], pix[4*i+2], pix[4*i+3] = 0xff, 0x59, 0x5e, 0xff
            default:
                pix[4*i], pix[4*i+1], pix[4*i+2], pix[4*i+3] = 0xff, 0xff, 0xff, 0xff
          }
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
    startTime := time.Now()
    rulesFilePath := "rules/sample03.txt"
    if len(os.Args) == 2 {
        rulesFilePath = os.Args[1]
    }

    file, err := os.Open(rulesFilePath)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    reader := bufio.NewReader(file)

    var neighborhoods []Neighborhood
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

                // Kind of scuffed
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
            if len(neighborhoods) == currentNeighborhood {
                neighborhoods = append(neighborhoods, Neighborhood{
                    neighbors: []Coordinate{},
                })
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
            if boolValue {
			    neigh := neighborhoods[currentNeighborhood]
                newX, newY := currentRow - dx, column - dy
                // skip middle cell
                if !(newX==0&&newY==0) {
                    neigh.neighbors = append(neigh.neighbors, Coordinate{x:newX, y:newY,})
                }
			    neighborhoods[currentNeighborhood] = neigh
            }
        }
    }

    fmt.Printf("Finished reading %s in: %s\n", rulesFilePath, time.Now().Sub(startTime))
    fmt.Printf("Loaded %d neighborhoods.\n", len(neighborhoods))
    fmt.Printf("Loaded %d rules.\n", len(rules))

    return EvolutionRules{
        numNeighborhoods: len(neighborhoods),
        neighborhoods: neighborhoods,
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
