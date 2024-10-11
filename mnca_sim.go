package main

import (
	"log"
    "bufio"
    "os"
    "fmt"
    "strconv"
    "strings"
    "math/rand/v2"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
    screenWidth = 512
    screenHeight = 384
)

/*  Rule
 *  @neighborhood   int corresponding to the neighborhood for the rule
 *  @min            float32 min value in interval
 *  @max            float32 max value in interval
 *  @nextState      bool next state if average is in this interval
 */
type Rule struct {
    neighborhood int
    min float32
    max float32
    nextState bool
}

// Returns where a value is inside a Rule's interval
func (i *Rule) Contains(val float32) bool {
    return val >= i.min && val <= i.max
}

// x y coordinate
type Coordinate struct {
    x int
    y int
}

/*  Neighborhood
 *  @validNeighbors int number of valid neighbors (len of neighbors[])
 *  @neighbors      []Coordinate of relative coordinates for neighboring cells
 */
type Neighborhood struct {
    validNeighbors int
    neighbors []Coordinate
}

/*  EvolutionRules
 *  @numNeighborhoods   int number of neighborhoods
 *  @neighborhoods      []Neighborhood
 *  @rulesList          []Rule
 */
type EvolutionRules struct {
    numNeighborhoods int
    neighborhoods []Neighborhood
    rulesList []Rule
}

/*  World
 *  @grid   []bool current state of World
 *  @width  int columns
 *  @height int rows
 *  @rules  EvolutionRules for World
 */
type World struct {
    grid []bool
    width int
    height int
    rules EvolutionRules
}

func InitializeWorld(width, height int) *World {
    w := &World {
        grid: make([]bool, width*height),
        width: width,
        height: height,
        rules: readNeighborhoods(),
    }

    for y := 0; y < w.height; y++ {
        for x := 0; x < w.width; x++ {
            w.grid[y*w.width+x] = rand.IntN(100) <= 60
        }
    }
    return w
}

func (w *World) Update() {
    width := w.width
    height := w.height
    next := make([]bool, width*height)

    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            sumAvgs := make([]float32, w.rules.numNeighborhoods)

            // calculate % of alive neighbors
            for i, _ := range sumAvgs {
                nCount := neighborCount(w.grid, w.width, w.height, x, y, &w.rules.neighborhoods[i])
                sumAvgs[i] = float32(nCount) / float32(w.rules.neighborhoods[i].validNeighbors)
            }

            // calculate next state based on EvolutionRules
            next[y*width+x] = w.grid[y*width+x]
            for _, interval := range w.rules.rulesList {
                if interval.Contains(sumAvgs[interval.neighborhood]) {
                    next[y*width+x] = interval.nextState
                }
            }
        }
    }

    // update grid with next state
    w.grid = next
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
            pix[4*i] = 0xff
            pix[4*i+1] = 0xff
            pix[4*i+2] = 0xff
            pix[4*i+3] = 0xff
        } else {
            pix[4*i] = 0
            pix[4*i+1] = 0
            pix[4*i+2] = 0
            pix[4*i+3] = 0
        }
    }
}

type Game struct {
    grid *World
    pixels []byte
}

func (g *Game) Update() error {
    g.grid.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    if g.pixels == nil {
        g.pixels = make([]byte, screenWidth*screenHeight*4)
    }
    g.grid.Draw(g.pixels)
    screen.WritePixels(g.pixels)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// this is probably not great
func readNeighborhoods() (EvolutionRules) {
    file, err := os.Open("rules.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    scanner.Scan()
    n, err := strconv.Atoi(scanner.Text())
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    r := EvolutionRules{
        numNeighborhoods: n,
        neighborhoods: make([]Neighborhood, n),
        rulesList: nil,
    }

    for i := 0; i < n; i++ {
        scanner.Scan()
        numNeighbors, err := strconv.Atoi(scanner.Text())
        if err != nil {
            log.Fatal(err)
        }

        defer file.Close()

        currentNeighborhood := Neighborhood{
            validNeighbors: numNeighbors,
            neighbors: make([]Coordinate, numNeighbors),
        }

        for j := 0; j < numNeighbors; j++ {
            scanner.Scan()
            line := scanner.Text()
            fields := strings.Fields(line)

            if len(fields) != 2 {
                fmt.Println("Invalid line format: ", line)
                continue
            }

            int1, err1 := strconv.Atoi(fields[0])
            int2, err2 := strconv.Atoi(fields[1])

            if err1 != nil || err2 != nil {
                fmt.Println("Error converting fields: ", fields)
                continue
            }

            coordinate := Coordinate{
                x: int1,
                y: int2,
            }

            currentNeighborhood.neighbors[j] = coordinate
        }
        r.neighborhoods[i] = currentNeighborhood
    }

    fmt.Printf("Successfully loaded %d neighborhoods\n", len(r.neighborhoods))
    scanner.Scan()
    rCount, err := strconv.Atoi(scanner.Text())
    if err != nil {
        fmt.Println("Could not read rule count")
    }
    defer file.Close()

    r.rulesList = make([]Rule, rCount)

    for i := 0; i < rCount; i++ {
        scanner.Scan()
        line := scanner.Text()
        fields := strings.Fields(line)

        if len(fields) != 4 {
            fmt.Println("Error reading line: ", line)
            continue
        }

        neighborhoodNum, err1 := strconv.Atoi(fields[0])
        intervalMin, err2 := strconv.ParseFloat(fields[1], 32)
        intervalMax, err3 := strconv.ParseFloat(fields[2], 32)
        nextState, err4 := strconv.ParseBool(fields[3])

        if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
            fmt.Println("error reading fields", fields)
            continue
        }

        interval := Rule{
            neighborhood: neighborhoodNum,
            min: float32(intervalMin),
            max: float32(intervalMax),
            nextState: nextState,
        }

        r.rulesList[i] = interval
    }

    fmt.Printf("Successfully loaded %d neighborhood rules\n", len(r.rulesList))
    return r
}

func main() {
    g := &Game {
        grid: InitializeWorld(screenWidth, screenHeight),
    }

	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("MNCA Simulator")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
