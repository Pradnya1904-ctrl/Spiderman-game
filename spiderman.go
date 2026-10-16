package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

type GameState int

const (
	StateStart GameState = iota
	StatePlaying
	StateGameOver
)

type GameObject struct {
	x, y          float32
	width, height float32
	speedY        float32
	active        bool
}

type Game struct {
	state       GameState
	playerX     float32
	playerY     float32
	webActive   bool
	webTargetX  float32
	webTargetY  float32
	score       int
	peopleSaved int

	gwen     GameObject
	people   []GameObject
	monsters []GameObject

	gwenSpawnTimer   int
	peopleSpawnTimer int
	monsterTimer     int
}

func NewGame() *Game {
	g := &Game{}
	g.Reset()
	return g
}

func (g *Game) Reset() {
	g.state = StateStart
	g.playerX = screenWidth / 2
	g.playerY = screenHeight - 60
	g.score = 0
	g.peopleSaved = 0
	g.webActive = false

	g.gwen = GameObject{active: false}
	g.people = nil
	g.monsters = nil

	g.gwenSpawnTimer = 0
	g.peopleSpawnTimer = 0
	g.monsterTimer = 0
}

func (g *Game) Update() error {
	rand.Seed(time.Now().UnixNano())

	switch g.state {
	case StateStart:
		// Game starts when Spider-Man throws a web (Left Mouse Click or Spacebar)
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.state = StatePlaying
			g.webActive = true
			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				cx, cy := ebiten.CursorPosition()
				g.webTargetX = float32(cx)
				g.webTargetY = float32(cy)
			} else {
				g.webTargetX = g.playerX
				g.webTargetY = 100
			}
		}

	case StatePlaying:
		// Move Player with Left/Right arrow keys or A/D
		if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
			if g.playerX > 20 {
				g.playerX -= 6
			}
		}
		if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
			if g.playerX < screenWidth-20 {
				g.playerX += 6
			}
		}

		// Web Action: Left Click shoots web to cursor position
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			cx, cy := ebiten.CursorPosition()
			g.webActive = true
			g.webTargetX = float32(cx)
			g.webTargetY = float32(cy)
		} else {
			g.webActive = false
		}

		// High Frequency Gwen Spawns (falls down frequently)
		g.gwenSpawnTimer++
		if !g.gwen.active && g.gwenSpawnTimer > 90 { // Spawns every ~1.5 seconds
			g.gwen = GameObject{
				x:      float32(rand.Intn(screenWidth - 40)),
				y:      0,
				width:  20,
				height: 25,
				speedY: 3.5, // High falling speed
				active: true,
			}
			g.gwenSpawnTimer = 0
		}

		// Update Gwen position
		if g.gwen.active {
			g.gwen.y += g.gwen.speedY

			// Check if saved by web or player proximity
			if g.webActive && g.checkWebHit(g.gwen) {
				g.gwen.active = false
				g.score += 50
			} else if g.checkPlayerHit(g.gwen) {
				g.gwen.active = false
				g.score += 50
			}

			// If Gwen falls past the bottom floor -> Game Over
			if g.gwen.y > screenHeight {
				g.state = StateGameOver
			}
		}

		// Spawn People needing rescue
		g.peopleSpawnTimer++
		if g.peopleSpawnTimer > 120 {
			g.people = append(g.people, GameObject{
				x:      float32(rand.Intn(screenWidth - 40)),
				y:      0,
				width:  20,
				height: 20,
				speedY: 1.8,
				active: true,
			})
			g.peopleSpawnTimer = 0
		}

		// Spawn Monsters
		g.monsterTimer++
		if g.monsterTimer > 150 {
			g.monsters = append(g.monsters, GameObject{
				x:      float32(rand.Intn(screenWidth - 40)),
				y:      0,
				width:  30,
				height: 30,
				speedY: 2.2,
				active: true,
			})
			g.monsterTimer = 0
		}

		// Update People
		for i := range g.people {
			if g.people[i].active {
				g.people[i].y += g.people[i].speedY

				// Save civilian with web or player contact
				if (g.webActive && g.checkWebHit(g.people[i])) || g.checkPlayerHit(g.people[i]) {
					g.people[i].active = false
					g.peopleSaved++
					g.score += 20
				}
			}
		}

		// Update Monsters & check interactions
		for i := range g.monsters {
			if g.monsters[i].active {
				g.monsters[i].y += g.monsters[i].speedY

				// Defeat monster with web
				if g.webActive && g.checkWebHit(g.monsters[i]) {
					g.monsters[i].active = false
					g.score += 30
				}
			}
		}

	case StateGameOver:
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			g.Reset()
		}
	}

	return nil
}

func (g *Game) checkWebHit(obj GameObject) bool {
	if !obj.active {
		return false
	}
	// Check distance from target of the web to object
	dx := g.webTargetX - (obj.x + obj.width/2)
	dy := g.webTargetY - (obj.y + obj.height/2)
	return (dx*dx + dy*dy) < 1200
}

func (g *Game) checkPlayerHit(obj GameObject) bool {
	if !obj.active {
		return false
	}
	return g.playerX < obj.x+obj.width &&
		g.playerX+30 > obj.x &&
		g.playerY < obj.y+obj.height &&
		g.playerY+30 > obj.y
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{20, 24, 40, 255})

	switch g.state {
	case StateStart:
		ebitenutil.DebugPrintAt(screen, "SPIDER-MAN: SAVE GWEN!", screenWidth/2-110, screenHeight/2-40)
		ebitenutil.DebugPrintAt(screen, "CLICK LEFT MOUSE BUTTON or PRESS SPACE to throw web & start!", screenWidth/2-210, screenHeight/2)

	case StatePlaying:
		// Draw Spider-Man (Red/Blue square)
		vector.DrawFilledRect(screen, g.playerX, g.playerY, 30, 30, color.RGBA{220, 30, 30, 255}, false)
		vector.DrawFilledRect(screen, g.playerX+5, g.playerY+20, 20, 10, color.RGBA{30, 30, 220, 255}, false)

		// Draw Web line when clicking
		if g.webActive {
			vector.StrokeLine(screen, g.playerX+15, g.playerY, g.webTargetX, g.webTargetY, 2, color.RGBA{240, 240, 240, 220}, false)
			vector.DrawFilledCircle(screen, g.webTargetX, g.webTargetY, 6, color.RGBA{255, 255, 255, 255}, false)
		}

		// Draw Gwen (Pink/Yellow square)
		if g.gwen.active {
			vector.DrawFilledRect(screen, g.gwen.x, g.gwen.y, g.gwen.width, g.gwen.height, color.RGBA{255, 105, 180, 255}, false)
			vector.DrawFilledRect(screen, g.gwen.x+2, g.gwen.y, g.gwen.width-4, 6, color.RGBA{255, 220, 0, 255}, false) // Blonde hair
		}

		// Draw People (Green)
		for _, p := range g.people {
			if p.active {
				vector.DrawFilledRect(screen, p.x, p.y, p.width, p.height, color.RGBA{40, 200, 80, 255}, false)
			}
		}

		// Draw Monsters (Purple/Dark)
		for _, m := range g.monsters {
			if m.active {
				vector.DrawFilledRect(screen, m.x, m.y, m.width, m.height, color.RGBA{140, 30, 180, 255}, false)
			}
		}

		// HUD
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d | People Saved: %d", g.score, g.peopleSaved), 10, 10)
		ebitenutil.DebugPrintAt(screen, "Controls: [A/D or Arrows] Move | [Left Click] Aim & Shoot Web", 10, 30)

	case StateGameOver:
		ebitenutil.DebugPrintAt(screen, "GAME OVER - Gwen fell down!", screenWidth/2-100, screenHeight/2-30)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Final Score: %d | Saved: %d", g.score, g.peopleSaved), screenWidth/2-100, screenHeight/2)
		ebitenutil.DebugPrintAt(screen, "Press 'R' to Restart", screenWidth/2-70, screenHeight/2+30)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Spider-Man: Save Gwen!")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
