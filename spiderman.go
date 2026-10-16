package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 600
	sampleRate   = 44100
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

	audioCtx      *audio.Context
	soundWebData  []byte
	soundSaveData []byte
	soundOverData []byte
}

func loadSound(path string) []byte {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Warning: Could not load sound %s: %v", path, err)
		return nil
	}
	return b
}

func (g *Game) playSound(data []byte) {
	if data == nil || g.audioCtx == nil {
		return
	}
	d, err := wav.DecodeWithoutResampling(bytes.NewReader(data))
	if err != nil {
		return
	}

	player, err := g.audioCtx.NewPlayer(d)
	if err != nil {
		return
	}
	player.Play()
}

func NewGame() *Game {
	g := &Game{}
	g.audioCtx = audio.NewContext(sampleRate)

	g.soundWebData = loadSound("assets/web.wav")
	g.soundSaveData = loadSound("assets/save.wav")
	g.soundOverData = loadSound("assets/gameover.wav")

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
	switch g.state {
	case StateStart:
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.state = StatePlaying
			g.webActive = true
			g.playSound(g.soundWebData)
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

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.playSound(g.soundWebData)
		}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			cx, cy := ebiten.CursorPosition()
			g.webActive = true
			g.webTargetX = float32(cx)
			g.webTargetY = float32(cy)
		} else {
			g.webActive = false
		}

		// Gwen Spawns
		g.gwenSpawnTimer++
		if !g.gwen.active && g.gwenSpawnTimer > 90 {
			g.gwen = GameObject{
				x:      float32(rand.Intn(screenWidth - 40)),
				y:      0,
				width:  20,
				height: 25,
				speedY: 3.5,
				active: true,
			}
			g.gwenSpawnTimer = 0
		}

		if g.gwen.active {
			g.gwen.y += g.gwen.speedY

			if (g.webActive && g.checkWebHit(g.gwen)) || g.checkPlayerHit(g.gwen) {
				g.gwen.active = false
				g.score += 50
				g.playSound(g.soundSaveData)
			}

			if g.gwen.y > screenHeight {
				g.state = StateGameOver
				g.playSound(g.soundOverData)
			}
		}

		// People Spawns
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

		for i := range g.people {
			if g.people[i].active {
				g.people[i].y += g.people[i].speedY
				if (g.webActive && g.checkWebHit(g.people[i])) || g.checkPlayerHit(g.people[i]) {
					g.people[i].active = false
					g.peopleSaved++
					g.score += 20
					g.playSound(g.soundSaveData)
				}
			}
		}

		// Monster Spawns
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

		for i := range g.monsters {
			if g.monsters[i].active {
				g.monsters[i].y += g.monsters[i].speedY
				if g.webActive && g.checkWebHit(g.monsters[i]) {
					g.monsters[i].active = false
					g.score += 30
					g.playSound(g.soundSaveData)
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
	screen.Fill(color.RGBA{20, 24, 40, 255})

	switch g.state {
	case StateStart:
		ebitenutil.DebugPrintAt(screen, "SPIDER-MAN: SAVE GWEN!", screenWidth/2-110, screenHeight/2-40)
		ebitenutil.DebugPrintAt(screen, "CLICK LEFT MOUSE BUTTON or PRESS SPACE to throw web & start!", screenWidth/2-210, screenHeight/2)

	case StatePlaying:
		vector.DrawFilledRect(screen, g.playerX, g.playerY, 30, 30, color.RGBA{220, 30, 30, 255}, false)
		vector.DrawFilledRect(screen, g.playerX+5, g.playerY+20, 20, 10, color.RGBA{30, 30, 220, 255}, false)

		if g.webActive {
			vector.StrokeLine(screen, g.playerX+15, g.playerY, g.webTargetX, g.webTargetY, 2, color.RGBA{240, 240, 240, 220}, false)
			vector.DrawFilledCircle(screen, g.webTargetX, g.webTargetY, 6, color.RGBA{255, 255, 255, 255}, false)
		}

		if g.gwen.active {
			vector.DrawFilledRect(screen, g.gwen.x, g.gwen.y, g.gwen.width, g.gwen.height, color.RGBA{255, 105, 180, 255}, false)
			vector.DrawFilledRect(screen, g.gwen.x+2, g.gwen.y, g.gwen.width-4, 6, color.RGBA{255, 220, 0, 255}, false)
		}

		for _, p := range g.people {
			if p.active {
				vector.DrawFilledRect(screen, p.x, p.y, p.width, p.height, color.RGBA{40, 200, 80, 255}, false)
			}
		}

		for _, m := range g.monsters {
			if m.active {
				vector.DrawFilledRect(screen, m.x, m.y, m.width, m.height, color.RGBA{140, 30, 180, 255}, false)
			}
		}

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
