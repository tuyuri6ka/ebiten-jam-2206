package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	screenWidth  = 320
	screenHeight = 320
	imageWidth   = 224
	imageHeight  = 224

	fontSize         = 10
	coefficient      = 1
	rotateNumMax     = 500
	fps          int = 1

	// gmae modes
	modeTitle  = 0
	modeGame   = 1
	modeFinish = 2

	defaultScore = 999999999
)

var (
	electromagnetImg *ebiten.Image
	arcadeFont       font.Face
)

//go:embed resources/images/electromagnet.png
var byteElectroMagnetImg []byte

// ebiten.Game interface を満たす型がEbitenには必要なので、
// この Game 構造体に Update, Draw, Layout 関数を持たせます。
type Game struct {
	count           int
	mode            int
	score           int
	hiscore         int
	angularVelocity int
	rotateNum       float64

	prevKey    ebiten.Key
	currentKey ebiten.Key
	keys       []ebiten.Key
}

// 構造体の初期化を行なっています。
func (g *Game) init() *Game {
	if g.hiscore == 0 {
		g.hiscore = defaultScore
	}
	g.count = 0
	g.score = 0
	g.angularVelocity = 0
	g.rotateNum = 0

	return g
}

// Update関数は、画面のリフレッシュレートに関わらず
// 常に毎秒60回呼ばれます（既定値）。
// 描画ではなく更新処理を行うことが推奨されます。
func (g *Game) Update() error {
	g.count++

	switch g.mode {
	case modeTitle:
		if g.isKeyJustPressed(ebiten.KeySpace) {
			g.mode = modeGame
		}
	case modeGame:
		if g.isKeyJustPressed(ebiten.KeyEscape) {
			g.mode = modeTitle
			g.init()
		}

		// チャージが満タンになったらゲームクリアになる
		g.rotateNum += float64(fps*g.angularVelocity) / 360
		if g.rotateNum >= float64(rotateNumMax) {
			// 記録の保存 端数は切り捨て
			g.score = g.count
			if g.score < g.hiscore {
				g.hiscore = g.score
			}
			g.mode = modeFinish
		}

		if g.isKeyJustPressed(ebiten.KeyArrowLeft) {
			if g.prevKey == ebiten.KeyArrowRight {
				g.angularVelocity += 1
			} else if g.prevKey == ebiten.KeyArrowLeft {
				g.angularVelocity -= 1
			}
			g.prevKey = ebiten.KeyArrowLeft
			g.currentKey = ebiten.KeyArrowLeft
		}
		if g.isKeyJustPressed(ebiten.KeyArrowRight) {
			if g.prevKey == ebiten.KeyArrowLeft {
				g.angularVelocity += 1
			} else if g.prevKey == ebiten.KeyArrowRight {
				g.angularVelocity -= 1
			}
			g.prevKey = ebiten.KeyArrowRight
			g.currentKey = ebiten.KeyArrowRight
		}
	case modeFinish:
		if g.isKeyJustPressed(ebiten.KeyEscape) {
			g.mode = modeTitle
			g.init()
		}
	}

	// キー入力をフレーム毎に受付
	g.keys = inpututil.AppendPressedKeys(g.keys[:0])

	return nil
}

// スペースキーが押されたかを判定しています
func (g *Game) isKeyJustPressed(key ebiten.Key) bool {
	if inpututil.IsKeyJustPressed(key) {
		return true
	}
	return false
}

func textDraw(g *Game, gauge string, charge float64, screen *ebiten.Image) {
	text.Draw(screen, fmt.Sprintf("gauge: %s", gauge), arcadeFont, 20, 10, color.Black)
	text.Draw(screen, fmt.Sprintf("velocity: %d", g.angularVelocity), arcadeFont, 20, 20, color.Black)

	if g.mode == modeGame {
		text.Draw(screen, fmt.Sprintf("score: %d", g.count), arcadeFont, 20, 30, color.Black)
	} else if g.mode == modeFinish {
		text.Draw(screen, fmt.Sprintf("score: %d", g.score), arcadeFont, 20, 30, color.Black)
		text.Draw(screen, fmt.Sprintf("%s", "Finish!!! \\(^o^)/"), arcadeFont, 20, 50, color.Black)
		text.Draw(screen, fmt.Sprintf("%s", "Restart game. Esc."), arcadeFont, 20, 300, color.Black)
	} else {
		// 何もしない
	}

	if g.hiscore < defaultScore {
		text.Draw(screen, fmt.Sprintf("hiscore: %d", g.hiscore), arcadeFont, 20, 40, color.Black)
	} else {
		// 何もしない
	}
}

func prepareDrawOption(g *Game) *ebiten.DrawImageOptions {
	// ebitenで画像を表示に関わるオプション設定をします
	option := &ebiten.DrawImageOptions{}

	// 画像の中心をスクリーンの左上に移動させる
	// ジオメトリマトリックス（回転や移動の処理）が適用される時の
	// 原点が画面の左上だから、加工のために中心に配置される画像を一旦原点に移動させる
	option.GeoM.Translate(-float64(imageWidth/2), -float64(imageHeight/2))

	// 構造体の状態を元に回転角度を算出する
	option.GeoM.Rotate(float64(float64((g.count*g.angularVelocity)%360) * 2 * math.Pi / 360))

	// 画像を拡大/縮小する
	option.GeoM.Scale(coefficient, coefficient)

	// 画像を好きな位置に移動させる
	// 今回は画像をスクリーンの中心に持ってくる
	option.GeoM.Translate(screenWidth/2, screenHeight/2)

	return option
}

// Draw関数は、画面のリフレッシュレートと同期して呼ばれます（既定値）。
// 描画処理のみを行うことが推奨されます。ここで状態の変更を行うといろいろ事故ります。
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)

	if g.mode == modeTitle {
		text.Draw(screen, fmt.Sprintf("Press Space! Game Start!"), arcadeFont, 20, int(screenHeight)/2-20, color.Black)
		text.Draw(screen, fmt.Sprintf("Press ← & → alternately!"), arcadeFont, 20, int(screenHeight)/2+0, color.Black)
		text.Draw(screen, fmt.Sprintf("The magnet will start spinning!"), arcadeFont, 0, int(screenHeight)/2+20, color.Black)
	} else if g.mode == modeGame {
		// ゲージの進捗度を計算する
		g.rotateNum += float64(fps*g.angularVelocity) / 360
		chargeStatus := int(g.rotateNum / 100)
		gauge := ""
		if g.rotateNum > rotateNumMax {
			gauge = "[" + strconv.Itoa(rotateNumMax) + "/" + strconv.Itoa(rotateNumMax) + "]"
			gauge += strings.Repeat("|", chargeStatus)
		} else if g.rotateNum >= 0 {
			gauge = "[" + strconv.Itoa(int(g.rotateNum)) + "/" + strconv.Itoa(rotateNumMax) + "]"
			gauge += strings.Repeat("|", chargeStatus)
		} else {
			gauge = "[" + strconv.Itoa(int(g.rotateNum)) + "/" + strconv.Itoa(rotateNumMax) + "]"
		}

		// テキストを画面に表示する
		textDraw(g, gauge, g.rotateNum, screen)

		// ebitenで画像を表示に関わるオプション設定をします
		option := prepareDrawOption(g)

		// オプションを元に画像を描画する
		screen.DrawImage(electromagnetImg, option)
	} else if g.mode == modeFinish {
		// ゲージの進捗度を計算する
		gauge := ""
		gauge = "[" + strconv.Itoa(rotateNumMax) + "/" + strconv.Itoa(rotateNumMax) + "]"
		gauge += strings.Repeat("|", rotateNumMax/100)

		// テキストを画面に表示する
		g.rotateNum += float64(fps*g.angularVelocity) / 360
		textDraw(g, gauge, g.rotateNum, screen)

		// ebitenで画像を表示に関わるオプション設定をします
		option := prepareDrawOption(g)

		// オプションを元に画像を描画する
		screen.DrawImage(electromagnetImg, option)
	}
}

// Layout関数は、ウィンドウのリサイズの挙動を決定します。画面サイズを返すのが無難だが適宜調整してください。
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// init関数はパッケージの初期化に使われる特殊な関数で、main関数が呼ばれる前に実行されます。
// ここではとりあえず画像ファイル、フォントデータを読み込むのに利用しています。
func init() {
	img, _, err := image.Decode(bytes.NewReader(byteElectroMagnetImg))
	if err != nil {
		log.Fatal(err)
	}
	electromagnetImg = ebiten.NewImageFromImage(img)

	tt, err := opentype.Parse(fonts.PressStart2P_ttf)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	arcadeFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}

// main関数から全ての処理が動き出します。ただしinit関数などの特殊な関数は除きます。
func main() {
	if err := _main(); err != nil {
		panic(err)
	}
}

func _main() error {
	g, err := newGame()
	if err != nil {
		return err
	}

	// ウィンドウズサイズとウィンドウ上部の表示タイトルを指定します。
	ebiten.SetWindowTitle("ElectroMagnetCharger (EbitengineGameJam202206)")
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	return ebiten.RunGame(g)
}

func newGame() (*Game, error) {
	// type struct
	g := &Game{}
	g.init()
	return g, nil
}
