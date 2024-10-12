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

	"github.com/hajimehoshi/ebiten/v2"
)

const (
    screenWidth = 512
    screenHeight = 384
)

type Rule struct {
    neighborhood int
    min, max float32
    nextState bool
}

// Returns where a value is inside a Rule's interval
func (r *Rule) Contains(val float32) bool {
    return val >= r.min && val <= r.max
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
        w.grid[i] = rand.IntN(100) < 60
    }

    return w
}

func (w *World) Update() {
    var wg sync.WaitGroup
    rowsPerRoutine := w.height/8

    for startY := 0; startY < w.height; startY += rowsPerRoutine {
        wg.Add(1)
        endY := startY + rowsPerRoutine
        if endY > w.height {
            endY = w.height
        }

        go func(start, end int) {
            defer wg.Done()
            for y := start; y < end; y++ {
                for x := 0; x < w.width; x++ {
                    sumAvgs := make([]float32, len(w.rules.neighborhoods))

                    // calculate % of alive neighbors
                    for i, neighborhood := range w.rules.neighborhoods {
                        nCount := neighborCount(w.grid, w.width, w.height, x, y, &neighborhood)
                        sumAvgs[i] = float32(nCount) / float32(neighborhood.validNeighbors)
                    }

                    // calculate next state based on EvolutionRules
                    nextState := w.grid[y*w.width+x]
                    for _, rule := range w.rules.rulesList {
                        if rule.Contains(sumAvgs[rule.neighborhood]) {
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

        // inbounds check
        if newX < 0 || newY < 0 || newX >= width || newY >= height {
            continue
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

func readNeighborhoods() (EvolutionRules) {
    file, err := os.Open("rules.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    scanner.Scan()
    numNeighborhoods, err := strconv.Atoi(scanner.Text())
    if err != nil {
        log.Fatal(err)
    }

    neighborhoods := make([]Neighborhood, numNeighborhoods)

    for i := 0; i < numNeighborhoods; i++ {
        scanner.Scan()
        numNeighbors, _ := strconv.Atoi(scanner.Text())

        neighbors := make([]Coordinate, numNeighbors)
        for j := 0; j < numNeighbors; j++ {
            scanner.Scan()
            fields := strings.Fields(scanner.Text())

            x, _ := strconv.Atoi(fields[0])
            y, _ := strconv.Atoi(fields[1])

            neighbors[j] = Coordinate{x, y}
        }
        neighborhoods[i] = Neighborhood{validNeighbors: numNeighbors, neighbors: neighbors}
    }

    fmt.Printf("Successfully loaded %d neighborhoods\n", len(neighborhoods))

    scanner.Scan()
    ruleCount, _ := strconv.Atoi(scanner.Text())

    rulesList := make([]Rule, ruleCount)

    for i := 0; i < ruleCount; i++ {
        scanner.Scan()
        fields := strings.Fields(scanner.Text())

        neighborhoodNum, _ := strconv.Atoi(fields[0])
        intervalMin, _ := strconv.ParseFloat(fields[1], 32)
        intervalMax, _ := strconv.ParseFloat(fields[2], 32)
        nextState, _ := strconv.ParseBool(fields[3])

        rulesList[i] = Rule{
            neighborhood: neighborhoodNum,
            min: float32(intervalMin),
            max: float32(intervalMax),
            nextState: nextState,
        }
    }

    fmt.Printf("Successfully loaded %d neighborhood rules\n", len(rulesList))
    return EvolutionRules{
        neighborhoods: neighborhoods,
        rulesList: rulesList,
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
