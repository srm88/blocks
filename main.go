package main

import (
    "fmt"
    glfw "github.com/go-gl/glfw3"
    glu "github.com/go-gl/glu"
    gl "github.com/go-gl/gl"
)

const (
    INCR_ANGLE float32 = 5
    INCR_SPACE float32 = 1
    UNIT float32 = 1.0
    HALF float32 = 0.49
    WORLD_X int = 50
    WORLD_Z int = 50
)
    
var (
    World Grid
    CursorX, CursorY, CursorZ int
    cursorBlock *Block
    POV *Camera
    SelectedCube int = 0
    CursorColor Color = MakeColor(1, 1, 1)
)

func init() {
    World = MakeGrid(WORLD_X, WORLD_Z)
    POV = &Camera{
        Loc{0, 0, -8},
        Orientation{30, 0, 315},
    }
    CursorX = 0
    CursorY = 0
    CursorZ = 0
    cursorBlock = &Block{Orientation{0,0,0},
                         MakeCube(UNIT/2.0, [6]Color{
                             CursorColor,
                             CursorColor,
                             CursorColor,
                             CursorColor,
                             CursorColor,
                             CursorColor,
                         })}
}

type Grid [][]*Block

func MakeGrid(x, y int) (g Grid) {
    g = make([][]*Block, x)
    for i := range g {
        g[i] = make([]*Block, y)
        for j := range g[i] {
            g[i][j] = MakeBlock()
        }
    }
    return
}

type Loc struct {
    X, Y, Z float32
}

type Orientation struct {
    Pitch, Roll, Yaw float32
}

type Camera struct {
    Loc
    Orientation
}

type Block struct {
    Orientation
    *Cube
}

func MakeBlock() *Block {
    return &Block{Orientation{0,0,0}, RegularCube()}
}

func errorCallback(err glfw.ErrorCode, desc string) {
    fmt.Printf("%v: %v\n", err, desc)
}

func copiedInit() {
    gl.ClearColor(0,0,0,0)
    gl.ShadeModel(gl.FLAT)
    gl.Enable(gl.CULL_FACE)
    gl.Viewport(320, 240, 640, 480)
    gl.MatrixMode(gl.PROJECTION)
    gl.LoadIdentity()
    gl.Frustum(-1, 1, -1, 1, 1.5, 20)
    gl.MatrixMode(gl.MODELVIEW)
}

type Quad struct {
    Vertices [4]int
    Color Color
}

type Color *[3]float32
func MakeColor(r, g, b float32) Color {
    return Color(&[3]float32{r, g, b})
}

type Vertex *[3]float32
func MakeVertex(x, y, z float32) Vertex {
    return Vertex(&[3]float32{x, y, z})
}


type Cube struct {
    Vertices [8]Vertex
    Quads [6]*Quad
}

func redraw() {
    gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
    gl.MatrixMode(gl.PROJECTION)
    gl.LoadIdentity()
    var viewport [4]int32
    gl.GetIntegerv(gl.VIEWPORT, viewport[:4])
    aspect := float64(viewport[2]) / float64(viewport[3])
    glu.Perspective(60, aspect, 1, 100)
    gl.MatrixMode(gl.MODELVIEW)
    gl.LoadIdentity()
    gl.Translatef(POV.X, POV.Y, POV.Z)
    gl.Rotatef(POV.Pitch, 1, 0, 0)
    gl.Rotatef(POV.Yaw, 0, 1, 0)
    gl.Rotatef(POV.Roll, 0, 0, 1)
    for x, row := range World {
        for z, block := range row {
            drawBlock(translateCoordinates(x, 0, z), block)
            if x == CursorX && z == CursorZ {
                drawCursor()
            }
        }
    }
}

func translateCoordinates(x, y, z int) Loc {
    return Loc{float32(x)*UNIT - float32(WORLD_X)/2.0,
               float32(y)*UNIT,
               float32(z)*UNIT - float32(WORLD_Z)/2.0}
}

func drawCursor() {
    drawBlock(translateCoordinates(CursorX, CursorY, CursorZ), cursorBlock)
}

func drawBlock(loc Loc, b *Block) {
    // Save current modelview matrix on to the stack
    gl.PushMatrix()
    // Move block 
    gl.Translatef(loc.X, loc.Y, loc.Z)
    // Rotate block
    gl.Rotatef(b.Pitch, 1, 0, 0)
    gl.Rotatef(b.Yaw, 0, 1, 0)
    gl.Rotatef(b.Roll, 0, 0, 1)
    // Start specifying the vertices of quads.
    gl.Begin(gl.QUADS)
    for _, quad := range b.Quads {
        gl.Color3fv(quad.Color)
        for _, vertex := range quad.Vertices {
            gl.Vertex3fv(b.Vertices[vertex])
        }
    }
    gl.End()
    // Restore old modelview matrix
    gl.PopMatrix()
}

func copiedReshape(window *glfw.Window, w int, h int) {
    gl.DrawBuffer(gl.FRONT_AND_BACK)
    //gl.Viewport(int(w/2), int(h/2), w, h)
    gl.MatrixMode(gl.PROJECTION)
    gl.LoadIdentity()
    gl.Frustum(-1, 1, -1, 1, 1.5, 20)
    gl.MatrixMode(gl.MODELVIEW)
}

func keyHandler(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
    if action == glfw.Release {
        switch key {
        // Panning
        case glfw.KeyH:
            POV.X -= INCR_SPACE
            break
        case glfw.KeyL:
            POV.X += INCR_SPACE
            break
        case glfw.KeyJ:
            if mods == glfw.ModShift {
                POV.Z -= INCR_SPACE
            } else {
                POV.Y -= INCR_SPACE
            }
            break
        case glfw.KeyK:
            if mods == glfw.ModShift {
                POV.Z += INCR_SPACE
            } else {
                POV.Y += INCR_SPACE
            }
            break
        // Cursor
        case glfw.KeyUp:
            if mods == glfw.ModShift {
                CursorY += 1
            } else {
                if CursorZ > 0 {
                    CursorZ -= 1
                }
            }
            break
        case glfw.KeyDown:
            if mods == glfw.ModShift {
                if CursorY > 0 {
                    CursorY -= 1
                }
            } else {
                if CursorZ < (WORLD_Z-1) {
                    CursorZ += 1
                }
            }
            break
        case glfw.KeyLeft:
            if CursorX > 0 {
                CursorX -= 1
            }
            break
        case glfw.KeyRight:
            if CursorX < (WORLD_X-1) {
                CursorX += 1
            }
            break
        }
    }
}

func main() {
    glfw.SetErrorCallback(errorCallback)

    if !glfw.Init() {
        panic("Can't init glfw")
    }
    defer glfw.Terminate()

    window, err := glfw.CreateWindow(640, 480, "Testing", nil, nil)
    if err != nil {
        panic(err)
    }

    window.MakeContextCurrent()
    window.SetSizeCallback(copiedReshape)
    window.SetKeyCallback(keyHandler)

    copiedInit()
    running := true
    for running && !window.ShouldClose() {
        //copiedDrawCube(angle)
        redraw()
        window.SwapBuffers()
        glfw.PollEvents()
        running = window.GetKey(glfw.KeyEscape) == glfw.Release
    }
}

func RegularCube() *Cube {
    return MakeCube(HALF, [6]Color{
        MakeColor(0.5, 0, 0),   // dark red
        MakeColor(1, 1, 0.3),   // yellow
        MakeColor(1, 0, 0),     // red
        MakeColor(0, 1, 0),     // green
        MakeColor(0.9, 0.5, 0), // orange
        MakeColor(0, 0, 1),     // bottom
    })
}

func MakeCube(size float32, colors [6]Color) *Cube {
    return &Cube{
        [8]Vertex{
            MakeVertex(-size, size, size),  //left,top,front
            MakeVertex(size, size, size),   //right,top,front
            MakeVertex(size, size, -size),  //right,top,back
            MakeVertex(-size, size, -size), //left,top,back
            MakeVertex(-size, -size, size), //left,bottom,front
            MakeVertex(size, -size, size),  //right,bottom,front
            MakeVertex(size, -size, -size), //right,bottom,back
            MakeVertex(-size, -size, -size), //left,bottom,back
        },
        [6]*Quad{
            &Quad{ // top
                [4]int{0, 1, 2, 3},
                colors[0],
            },
            &Quad{ // left
                [4]int{0, 3, 7, 4},
                colors[1],
            },
            &Quad{ // back
                [4]int{3, 2, 6, 7},
                colors[2],
            },
            &Quad{ // right
                [4]int{2, 1, 5, 6},
                colors[3],
            },
            &Quad{ // front
                [4]int{0, 4, 5, 1},
                colors[4],
            },
            &Quad{ // bottom
                [4]int{4, 7, 6, 5},
                colors[5],
            },
        },
    }
}
