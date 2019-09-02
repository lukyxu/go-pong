package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"
)

func main() {
	pixelgl.Run(run)
}

const (
	friction       = 0.65
	fps            = 60
	paddleAccel    = 0.7
	windowWidth    = 420
	windowHeight   = 600
	maxBounceAngle = math.Pi * 5 / 12
)

type paddle struct {
	xPos      float64
	yPos      float64
	yVel      float64
	maxVel    float64
	upAccel   bool
	downAccel bool
	width     int
	height    int
}

type ball struct {
	xPos   float64
	yPos   float64
	radius float64
	xVel   float64
	yVel   float64
}

type pongGame struct {
	player1  paddle
	player2  paddle
	gameBall ball
	paused   bool
	p1Score  int
	p2Score  int
}

func createPaddle(player int) paddle {
	outputPaddle := paddle{
		yPos:   260,
		width:  10,
		height: 80,
		yVel:   0,
	}
	if player == 1 {
		outputPaddle.xPos = 20
		outputPaddle.maxVel = 10
	} else {
		outputPaddle.xPos = 390
		outputPaddle.maxVel = 4
	}
	return outputPaddle
}

func (b ball) draw(win *pixelgl.Window) {
	imd := imdraw.New(nil)
	imd.Color = pixel.RGB(1, 1, 1)
	imd.Push(pixel.V(b.xPos, b.yPos))
	imd.Circle(b.radius, 0)
	imd.Draw(win)
}

func (p paddle) draw(win *pixelgl.Window) {
	imd := imdraw.New(nil)
	imd.Color = pixel.RGB(1, 1, 1)
	imd.Push(pixel.V(p.xPos, p.yPos), pixel.V(p.xPos+float64(p.width), p.yPos+float64(p.height)))
	imd.Rectangle(0)
	imd.Draw(win)
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pong",
		Bounds: pixel.R(0, 0, windowWidth, windowHeight),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	for !win.Closed() {
		game := pongGame{}
		game.p1Score = 0
		game.p2Score = 0
		resetGame(&game)
		for game.p1Score < 10 && game.p2Score < 10 {
			prevTime := time.Now()
			if win.Closed() {
				break
			}
			handleIO(win, &game)
			if !game.paused {
				updateGame(win, &game)
			}
			win.Clear(color.Black)

			atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)

			if game.p1Score == 0 && game.p2Score == 0 && game.paused {
				txt := text.New(pixel.V(windowWidth/2, windowHeight), atlas)
				txt.Dot.Y -= txt.BoundsOf("To play use the up and down arrow keys").H()
				txt.Dot.X -= txt.BoundsOf("To play use the up and down arrow keys").W() / 2
				fmt.Fprintf(txt, "To play use the up and down arrow keys")
				txt.Dot.Y -= txt.BoundsOf("First player to 10 point wins").H()
				txt.Dot.X = windowWidth/2 - txt.BoundsOf("First player to 10 point wins").W()/2
				fmt.Fprintf(txt, "First player to 10 point wins")
				txt.Draw(win, pixel.IM.Scaled(txt.Orig, 1.5))
			}

			// Draws the player scores onto the window.
			basicTxt := text.New(pixel.V(0, 0), atlas)
			fmt.Fprintln(basicTxt, "P1", game.p1Score)
			basicTxt.Draw(win, pixel.IM.Scaled(basicTxt.Orig, 4))
			basicTxt = text.New(pixel.V(windowWidth, 0), atlas)
			basicTxt.Dot.X -= basicTxt.BoundsOf("P2 " + strconv.Itoa(game.p2Score)).W()
			fmt.Fprintln(basicTxt, "P2", game.p2Score)
			basicTxt.Draw(win, pixel.IM.Scaled(basicTxt.Orig, 4))

			// Draws the game onto the window.
			game.gameBall.draw(win)
			game.player1.draw(win)
			game.player2.draw(win)

			if game.p1Score == 10 {
				txt := text.New(pixel.V(windowWidth/2, windowHeight), atlas)
				txt.Dot.Y -= txt.BoundsOf("Player 1 wins").H()
				txt.Dot.X -= txt.BoundsOf("Player 1 wins").W() / 2
				fmt.Fprintf(txt, "Player 1 wins")
				txt.Draw(win, pixel.IM.Scaled(txt.Orig, 4))
				win.Update()
				time.Sleep(time.Second * 2)
			} else if game.p2Score == 10 {
				fmt.Println("nani")
				txt := text.New(pixel.V(windowWidth/2, windowHeight), atlas)
				txt.Dot.Y -= txt.BoundsOf("Player 2 wins").H()
				txt.Dot.X -= txt.BoundsOf("Player 2 wins").W() / 2
				fmt.Fprintf(txt, "Player 2 wins")
				txt.Draw(win, pixel.IM.Scaled(txt.Orig, 4))
				win.Update()
				time.Sleep(time.Second * 2)
			} else {
				win.Update()
			}

			dt := time.Now().Sub(prevTime)
			delay := time.Second/fps - dt
			if delay > 0 {
				gameTimer := time.NewTicker(delay)
				<-gameTimer.C
			}
		}
	}
}

func handleIO(win *pixelgl.Window, game *pongGame) {
	if win.JustPressed(pixelgl.KeyUp) {
		game.player1.upAccel = true
		game.paused = false
	} else if win.JustPressed(pixelgl.KeyDown) {
		game.player1.downAccel = true
		game.paused = false
	}

	if win.JustReleased(pixelgl.KeyUp) {
		game.player1.upAccel = false
	} else if win.JustReleased(pixelgl.KeyDown) {
		game.player1.downAccel = false
	}
}

func updateGame(win *pixelgl.Window, game *pongGame) {
	// Handles player 1 paddle movements.
	if game.player1.upAccel {
		game.player1.yVel += paddleAccel
	} else if game.player1.downAccel {
		game.player1.yVel -= paddleAccel
	} else if (!game.player1.upAccel) && (!game.player1.downAccel) {
		game.player1.yVel *= friction
	}

	// Stops player 1 paddle from going over its max velocity.
	if game.player1.yVel > game.player1.maxVel {
		game.player1.yVel = game.player1.maxVel
	} else if game.player1.yVel < -game.player1.maxVel {
		game.player1.yVel = -game.player1.maxVel
	}

	// Stops player 1 paddle from going off screen.
	if game.player1.yPos < float64(-game.player1.height)/2 {
		game.player1.yPos = float64(-game.player1.height) / 2
	} else if game.player1.yPos > windowHeight-(float64(game.player1.height)/2) {
		game.player1.yPos = windowHeight - (float64(game.player1.height) / 2)
	}
	game.player1.yPos += game.player1.yVel

	// Handles compmuter player 2 paddle movements.
	if math.Abs(game.player2.yPos+float64(game.player2.height/2)-game.gameBall.yPos) < 15 {
		game.player2.yPos += (game.gameBall.yPos - (game.player2.yPos + float64(game.player2.height/2))) / float64(game.player2.height/2) * 4
	} else if game.player2.yPos+float64(game.player2.height/2) < game.gameBall.yPos {
		game.player2.yPos += game.player2.maxVel
	} else if game.player2.yPos+float64(game.player2.height/2) > game.gameBall.yPos {
		game.player2.yPos += -game.player2.maxVel
	}

	// Paddle collision with the ball.
	p1Collision := game.player1.xPos + float64(game.player1.width) + game.gameBall.radius
	p2Collision := game.player2.xPos - game.gameBall.radius
	if game.gameBall.xPos <= p1Collision && game.gameBall.xPos >= p1Collision-math.Abs(game.gameBall.xVel) &&
		game.gameBall.yPos <= game.player1.yPos+float64(game.player1.height) &&
		game.gameBall.yPos >= game.player1.yPos {
		game.gameBall.xPos = p1Collision
		relativeYIntersect := game.gameBall.yPos - (game.player1.yPos + float64(game.player1.height)/2)
		normalizedYIntersect := relativeYIntersect / (float64(game.player1.height) / 2)
		bounceAngle := normalizedYIntersect * maxBounceAngle
		speed := math.Pow(math.Pow(game.gameBall.xVel, 2)+math.Pow(game.gameBall.yVel, 2), 0.5)
		game.gameBall.xVel = math.Cos(bounceAngle) * speed
		game.gameBall.yVel = math.Sin(bounceAngle) * speed
	} else if game.gameBall.xPos >= p2Collision && game.gameBall.xPos <= p2Collision+game.gameBall.xVel &&
		game.gameBall.yPos <= game.player2.yPos+float64(game.player2.height) &&
		game.gameBall.yPos >= game.player2.yPos {
		game.gameBall.xPos = p2Collision
		relativeYIntersect := game.gameBall.yPos - (game.player2.yPos + float64(game.player2.height)/2)
		normalizedYIntersect := relativeYIntersect / (float64(game.player2.height) / 2)
		bounceAngle := normalizedYIntersect * maxBounceAngle
		speed := math.Pow(math.Pow(game.gameBall.xVel, 2)+math.Pow(game.gameBall.yVel, 2), 0.5)
		game.gameBall.xVel = -math.Cos(bounceAngle) * speed
		game.gameBall.yVel = math.Sin(bounceAngle) * speed
	}

	// Ball bouncing off the top and bottom walls respectively.
	if game.gameBall.yPos <= game.gameBall.radius || game.gameBall.yPos >= windowHeight-game.gameBall.radius {
		game.gameBall.yVel = -game.gameBall.yVel
	}

	game.gameBall.xPos += game.gameBall.xVel
	game.gameBall.yPos += game.gameBall.yVel

	// Checks if the game has ended and increases the score of the winner.
	if game.gameBall.xPos < -20 {
		time.Sleep(500 * time.Millisecond)
		game.p2Score++
		resetGame(game)
	} else if game.gameBall.xPos > 450 {
		time.Sleep(500 * time.Millisecond)
		game.p1Score++
		resetGame(game)
	}
}

func resetGame(g *pongGame) {
	g.player1 = createPaddle(1)
	g.player2 = createPaddle(2)
	g.gameBall = ball{xPos: 210, yPos: 300, radius: 5, xVel: randVelocity() * 2, yVel: randVelocity() * 0.5}
	g.paused = true
	if g.gameBall.xVel < 0 {
		g.gameBall.xVel = -g.gameBall.xVel
	}
}

func randVelocity() float64 {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	output := (rng.Float64() * 1.5) + 3.5
	if rng.Intn(2) == 1 {
		output *= -1
	}
	return output
}

func array(xs [3]int) {

}
