package main

import (

	"math/rand"
	"log"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
	"golang.org/x/mobile/exp/gl/glutil"

	"encoding/binary"
	"golang.org/x/mobile/exp/f32"

	"time"
	"golang.org/x/mobile/exp/sprite/clock"

	"image"
	"fmt"
	"golang.org/x/mobile/geom"
	"image/color"
	"github.com/golang/freetype/truetype"


)

var (
	startTime = time.Now()

	ticker = time.NewTicker(time.Millisecond * 500)
	counter = 0
	destroyed = 0
	
	game 	 *Game
	program  gl.Program
	buf      gl.Buffer

	images    *glutil.Images
	fonts       *truetype.Font

	position gl.Attrib
	offset   gl.Uniform
	colorShader    gl.Uniform

	touchX float32
	touchY float32

	fieldsX [elements]float32
	fieldsY [elements]float32
	toDraw 	[elements]bool
	visible [elements]bool
	timer float32;
)

func main() {
	app.Main(func(a app.App) {
		var glctx gl.Context
		var sz size.Event
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx, _ = e.DrawContext.(gl.Context)
					onStart(glctx)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					onStop(glctx)
					glctx = nil
				}
			case size.Event:
				sz = e
			case paint.Event:
				if glctx == nil || e.External {
					continue
				}
				onPaint(glctx, sz)
				a.Publish()
				a.Send(paint.Event{}) // keep animating
			case touch.Event:
				// if down := e.Type == touch.TypeBegin; down || e.Type == touch.TypeEnd {
				// 	game.Press(down)
				// }
				touchX = e.X
				touchY = e.Y
				checkIfTouched(e.X,e.Y)
			}
		}
	})
}

func checkIfTouched(x float32, y float32) {
	xSize := 35
	ySize := 110
	for i := range fieldsY {
		if((x > fieldsX[i] && x < fieldsX[i] + float32(xSize) ) && (y < fieldsY[i] && y > fieldsY[i] - float32(ySize) ) ){
			toDraw[i] = false
			
		}	
	}
}

func onStart(glctx gl.Context) {
	var err error
	program, err = glutil.CreateProgram(glctx, vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	fonts, err = LoadCustomFont()
	if err != nil {
		log.Fatalf("error parsing font: %v", err)
	}

	images = glutil.NewImages(glctx)
	buf = glctx.CreateBuffer()
	glctx.BindBuffer(gl.ARRAY_BUFFER, buf)
	glctx.BufferData(gl.ARRAY_BUFFER, triangleData, gl.STATIC_DRAW)
	//glctx.BufferData(gl.ARRAY_BUFFER, triangleData2, gl.STATIC_DRAW)

	position = glctx.GetAttribLocation(program, "position")
	colorShader = glctx.GetUniformLocation(program, "color")
	offset = glctx.GetUniformLocation(program, "offset")

	touchY = 500
	touchX = 500
 
	go func () {
		for t := range ticker.C {
			visible[counter] = true;
			counter += 1;
			log.Printf("%v",t)
		}	
	}()

	for i:= range fieldsY {
		fieldsY[i] = 150 + rand.Float32()*(screenSizeY - 150)
		fieldsX[i] = rand.Float32()*screenSizeX
		toDraw[i] = true;
		visible[i] = false;
	}
}

var triangleData = f32.Bytes(binary.LittleEndian,
	0.0, 0.1, 0.0, // top left
	0.0, 0.0, 0.0, // bottom left
	0.1, 0.0, 0.0, // bottom right
)

var triangleData2 = f32.Bytes(binary.LittleEndian,
	0.4, 0.4, 0.0, // top left
	0.0, 0.4, 0.0, // bottom left
	0.4, 0.0, 0.0, // bottom right
)

const vertexShader = `#version 100
uniform vec2 offset;
attribute vec4 position;
void main() {
	// offset comes in with x/y values between 0 and 1.
	// position bounds are -1 to 1.
	vec4 offset4 = vec4(2.0*offset.x-1.0, 1.0-2.0*offset.y, 0, 0);
	gl_Position = position+ offset4;
}`

const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
	gl_FragColor = color;
}`

func onStop(glctx gl.Context) {
	images.Release()
	glctx.DeleteProgram(program)
	glctx.DeleteBuffer(buf)
}

func onPaint(glctx gl.Context, sz size.Event) {
	glctx.ClearColor(1, 0, 0, 1)
	glctx.Clear(gl.COLOR_BUFFER_BIT)

	glctx.UseProgram(program)

	glctx.Uniform4f(colorShader, 0, 1, 0, 1)
	// var x float32
	// var y float32

	// timestamp := time.Since(startTime)
	// now := timestamp/time.Second

	if (counter == elements) {
		ticker.Stop()
	}
	// timer += 0.01
	// if (timer/1 > 1) {
	// x = rand.Float32()*500.0
	// y = rand.Float32()*500.0
	// timer = 0;
	// }

	

	for i:= range fieldsY{
		if(toDraw[i] == true && visible[i] == true) {
			glctx.Uniform2f(offset, fieldsX[i]/float32(sz.WidthPx), fieldsY[i]/float32(sz.HeightPx))


			glctx.BindBuffer(gl.ARRAY_BUFFER, buf)
			glctx.EnableVertexAttribArray(position)
			glctx.VertexAttribPointer(position, coordsPerVertex, gl.FLOAT, false, 0, 0)
			glctx.DrawArrays(gl.TRIANGLES, 0, vertexCount)
			glctx.DisableVertexAttribArray(position)
		}
		if(toDraw[i] == false) {
			destroyed += 1
		}
	}

	renderText(sz,glctx,images)
}

 func renderText(sz size.Event, glctx gl.Context, images *glutil.Images) {
	//headerHeightPx, footerHeightPx := 100, 100

	now := clock.Time(time.Since(startTime) / time.Second)
	
	loading := &TextSprite{
		text:            fmt.Sprintf("Spawned : %d Destroyed : %d Time Elapsed: %d", counter, destroyed, int(now)),
		font:            fonts,
		widthPx:         sz.WidthPx,
		heightPx:        200,
		textColor:       image.White,
		backgroundColor: image.NewUniform(color.RGBA{0x00, 0x00, 0xFF, 0xFF}),
		fontSize:        24,
		xPt:             0,
		yPt:             0,
		align:           Left,
	}
	loading.Render(sz)
	destroyed = 0
 }

const (
	elements = 300
	screenSizeX = 1080
	screenSizeY = 1920
	coordsPerVertex = 3
	vertexCount     = 3
)


func PxToPt(sz size.Event, sizePx int) geom.Pt {
	return geom.Pt(float32(sizePx) / sz.PixelsPerPt)
}
