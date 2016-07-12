// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"time"
	"fmt"

	"os"
	"path/filepath"
	"strings"

	"os/exec"
	"log"

	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/gxfont"
	"github.com/google/gxui/math"
	"github.com/google/gxui/samples/flags"

	"github.com/google/gxui/samples/file_dlg/roots"
)

import "github.com/Code-hex/pget"
import "github.com/atotto/clipboard"
//////////
var (
	//fileColor      = gxui.Color{R: 0.7, G: 0.8, B: 1.0, A: 1}
	//directoryColor = gxui.Color{R: 0.8, G: 1.0, B: 0.7, A: 1}
	fileColor      = gxui.Color{R: 0.0, G: 0.0, B: 1.0, A: 1}
	directoryColor = gxui.Color{R: 0.0, G: 0.0, B: 0.0, A: 1}
)
// filesAt returns a list of all immediate files in the given directory path.
/*func (f []string) removeHiddenFiles() []string {

	for i := 0; i < len(f); i++ {
		fileName := f[i]
		linx := strings.LastIndex(fileName,"/")
		// remove from subdirs if contains "." 
		if strings.Index(fileName[linx:], ".") !=-1 {
			f = append( f[:i],f[i+1:]...)
			i -=1
		}
	}
	return f
}
*/
func filesAt(path string) []string {
	files := []string{}
	filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err == nil && path != subpath {
			files = append(files, subpath)
			if info.IsDir() {
				return filepath.SkipDir
			}
		}
		return nil
	})
	//files = files.removeHiddenFiles()

	//
	for i := 0; i < len(files); i++ {
		fileName := files[i]
		linx := strings.LastIndex(fileName,"/")
		// remove from subdirs if contains "." 
		if strings.Index(fileName[linx:], ".") !=-1 {
			files = append( files[:i],files[i+1:]...)
			i -=1
		}
	}
	//
	return files
}
// filesAdapter is an implementation of the gxui.ListAdapter interface.
// The AdapterItems returned by this adapter are absolute file path strings.
type filesAdapter struct {
	gxui.AdapterBase
	files []string // The absolute file paths
}

// SetFiles assigns the specified list of absolute-path files to this adapter.
func (a *filesAdapter) SetFiles(files []string) {
	a.files = files
	a.DataChanged(false)
}

func (a *filesAdapter) Count() int {
	return len(a.files)
}

func (a *filesAdapter) ItemAt(index int) gxui.AdapterItem {
	return a.files[index]
}

func (a *filesAdapter) ItemIndex(item gxui.AdapterItem) int {
	path := item.(string)
	for i, f := range a.files {
		if f == path {
			return i
		}
	}
	return -1 // Not found
}

func (a *filesAdapter) Create(theme gxui.Theme, index int) gxui.Control {
	path := a.files[index]
	_, name := filepath.Split(path)
	label := theme.CreateLabel()
	label.SetText(name)
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		label.SetColor(directoryColor)
	} else {
		label.SetColor(fileColor)
	}
	return label
}

func (a *filesAdapter) Size(gxui.Theme) math.Size {
	return math.Size{W: math.MaxSize.W, H: 20}
}

// directory implements the gxui.TreeNode interface to represent a directory
// node in a file-system.
type directory struct {
	path    string   // The absolute path of this directory.
	subdirs []string // The absolute paths of all immediate sub-directories.
}

func (d directory) removeHiddenDirs() directory {

	for i := 0; i < len(d.subdirs); i++ {
		dirName := d.subdirs[i]
		linx := strings.LastIndex(dirName,"/")
		// remove from subdirs if contains "." 
		if strings.Index(dirName[linx:], ".") !=-1 {
			d.subdirs = append( d.subdirs[:i],d.subdirs[i+1:]...)
			i -=1
		}
	}
	return d
}

// directoryAt returns a directory structure populated with the immediate
// subdirectories at the given path.
func directoryAt(path string) directory {
	directory := directory{path: path}
	filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err == nil && path != subpath {
			if info.IsDir() {
				directory.subdirs = append(directory.subdirs, subpath)
				//CHANGED
				//directory = directory.removeHiddenDirs()
				return filepath.SkipDir
			}
		}
		return nil
	})
	directory =directory.removeHiddenDirs()
	return directory
}

// Count implements gxui.TreeNodeContainer.
func (d directory) Count() int {
	return len(d.subdirs)
}

// NodeAt implements gxui.TreeNodeContainer.
func (d directory) NodeAt(index int) gxui.TreeNode {
	return directoryAt(d.subdirs[index])
}

// ItemIndex implements gxui.TreeNodeContainer.
func (d directory) ItemIndex(item gxui.AdapterItem) int {
	path := item.(string)
	if !strings.HasSuffix(path, string(filepath.Separator)) {
		path += string(filepath.Separator)
	}
	for i, subpath := range d.subdirs {
		subpath += string(filepath.Separator)
		if strings.HasPrefix(path, subpath) {
			return i
		}
	}
	return -1
}

// Item implements gxui.TreeNode.
func (d directory) Item() gxui.AdapterItem {
	return d.path
}

// Create implements gxui.TreeNode.
func (d directory) Create(theme gxui.Theme) gxui.Control {
	_, name := filepath.Split(d.path)
	if name == "" {
		name = d.path
	}
	l := theme.CreateLabel()
	l.SetText(name)
	l.SetColor(directoryColor)
	return l
}

// directoryAdapter is an implementation of the gxui.TreeAdapter interface.
// The AdapterItems returned by this adapter are absolute file path strings.
type directoryAdapter struct {
	gxui.AdapterBase
	directory
}


func (a directoryAdapter) Size(gxui.Theme) math.Size {
	return math.Size{W: math.MaxSize.W, H: 20}
}

// Override directory.Create so that the full root is shown, unaltered.
func (a directoryAdapter) Create(theme gxui.Theme, index int) gxui.Control {
	l := theme.CreateLabel()
	l.SetText(a.subdirs[index])
	l.SetColor(directoryColor)
	return l
}
func DirSize(path string) (int64, error) {
    var size int64
    err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
        if !info.IsDir() {
            size += info.Size()
        }
        return err
    })
    return size, err
}
//////////
//////////
//////////
//////////
//////////
//////////
//////////
//////////
//////////
//////////
//////////
//////////
//////////
//////////

func appMain(driver gxui.Driver) {
	//
	p := pget.New()
	//
	theme := flags.CreateTheme(driver)

	window := theme.CreateWindow(800, 460, "Paraload")
	window.SetBackgroundBrush(gxui.CreateBrush(gxui.White))
	window.SetScale(flags.DefaultScaleFactor)

	urlBox := theme.CreateTextBox()
	urlBox.SetDesiredWidth(math.MaxSize.W)
	urlBox.SetText("Please paste URL for download")
	
	


	//////////////////
	fullpath := theme.CreateTextBox()
	fullpath.SetDesiredWidth(math.MaxSize.W)

	//
	downloadButton := theme.CreateButton()
	downloadButton.SetText("Download")
	downloadButton.OnClick(func(gxui.MouseEvent) {
		fmt.Printf("url is  '%s' \n", urlBox.Text())
		fmt.Printf("path is  '%s' \n", fullpath.Text())
		//window.Close()
		

		//
		p.Run(urlBox.Text(),fullpath.Text()+"/")

		dn := p.DirName()
		fmt.Printf("dn is---  '%s' \n", dn)
		dSize, err := DirSize(dn)
		fmt.Printf("dSize  '%s' \n", fullpath.Text())

		//

	})

	directories := theme.CreateTree()


	dd := directory{subdirs:roots.Roots(),}
	/*for i := 0; i < len(dd.subdirs); i++ {
		fmt.Printf("sd------- '%s' \n", dd.subdirs[i])
	}*/
	dd=dd.removeHiddenDirs()
	/*for i := 0; i < len(dd.subdirs); i++ {
		fmt.Printf("-----sd------- '%s' \n", dd.subdirs[i])
	}*/
	/*directories.SetAdapter(&directoryAdapter{
		directory: directory{
			subdirs: roots.Roots(),
		},
	})
	*/
	
	directories.SetAdapter(&directoryAdapter{
		directory: dd,
	})

	// filesAdapter is the adapter used to show the currently selected directory's
	// content. The adapter has its data changed whenever the selected directory
	// changes.
	filesAdapter := &filesAdapter{}
	files := theme.CreateList()
	files.SetAdapter(filesAdapter)

	open := theme.CreateButton()
	open.SetText("Paste URL")
	open.OnClick(func(gxui.MouseEvent) {
		fmt.Printf("paste clicked-------  \n")
		if text,err := clipboard.ReadAll();err!=nil {
			fmt.Printf("ERROR------- '%s' \n", err)

		} else{
			fmt.Printf("text------- '%s' \n", text)
			urlBox.SetText(text)}
		
		//fmt.Printf("File '%s' selected!\n", files.Selected())
		//window.Close()
	})
	//open.OnClick(func(gxui.MouseEvent) {
	//	fmt.Printf("File '%s' selected!\n", files.Selected())
	//	window.Close()
	//})
	// If the user hits the enter key while the fullpath control has focus,
	// attempt to select the directory.
	fullpath.OnKeyDown(func(ev gxui.KeyboardEvent) {
		if ev.Key == gxui.KeyEnter || ev.Key == gxui.KeyKpEnter {
			path := fullpath.Text()
			if directories.Select(path) {
				directories.Show(path)

			}
		}
	})
	// When the directory selection changes, update the files list
	directories.OnSelectionChanged(func(item gxui.AdapterItem) {
		
		dir := item.(string)
		//
		//fmt.Printf("CHEEEEECK------- '%s' \n", dir)
		//

		filesAdapter.SetFiles(filesAt(dir))
		fullpath.SetText(dir)
	})

	// When the file selection changes, update the fullpath text
	files.OnSelectionChanged(func(item gxui.AdapterItem) {
		fullpath.SetText(item.(string))
	})
	// When the user double-clicks a directory in the file list, select it in the
	// directories tree view.
	files.OnDoubleClick(func(gxui.MouseEvent) {
		if path, ok := files.Selected().(string); ok {
			if fi, err := os.Stat(path); err == nil && fi.IsDir() {
				if directories.Select(path) {
					directories.Show(path)
				}
			} else {
				fmt.Printf("File '%s' selected!\n", path)
				window.Close()
			}
		}
	})
	// Start with the CWD selected and visible.
	if cwd, err := os.Getwd(); err == nil {
		if directories.Select(cwd) {
			directories.Show(directories.Selected())
			//fmt.Printf(directories.Selected())
		}
	}
	//
	openFolderButton := theme.CreateButton()
	openFolderButton.SetText("Open Folder")
	openFolderButton.OnClick(func(gxui.MouseEvent) {
		//fmt.Printf("url is  '%s' \n", openFolderButton.Text())
		dirr:=fullpath.Text()

		cmd := exec.Command("open", dirr)
		err := cmd.Start()

		if err != nil {
			log.Fatal(err)
		}
		//log.Printf("Waiting for command to finish...")
		err = cmd.Wait()
		//log.Printf("Command finished with error: %v", err)
		

	})
	//////////////////

	font, err := driver.CreateFont(gxfont.Default, 75)
	if err != nil {
		panic(err)
	}

	

	label := theme.CreateLabel()
	label.SetFont(font)
	label.SetText("Hello world")
	//PROGRESS BAR
	//
	//p.Utils.subDirSize(p.Utils.dirname)
	
	
	//dSize = int(dSize.(int32))

	
	//if dSize, err := DirSize(p.DirName());err!=nil {
	//	fmt.Printf("ERR---  '%s' \n",err)
	//}
	

	//fSize := p.FileSize()
	//
	
	progressBar := theme.CreateProgressBar()
	progressBar.SetDesiredSize(math.Size{W: 800, H: 20})
	progressBar.SetTarget(100)


	progress := 0
	pause := time.Millisecond * 500
	var timer *time.Timer
	timer = time.AfterFunc(pause, func() {
		driver.Call(func() {
			//progress = (progress + 3) % progressBar.Target()
			progress = (progress + 3) % progressBar.Target()
			progressBar.SetProgress(progress)
			timer.Reset(pause)
		})
	})
	//window.AddChild(label)

	ticker := time.NewTicker(time.Millisecond * 30)
	go func() {
		phase := float32(0)
		for _ = range ticker.C {
			c := gxui.Color{
				R: 0.75 + 0.25*math.Cosf((phase+0.000)*math.TwoPi),
				G: 0.75 + 0.25*math.Cosf((phase+0.333)*math.TwoPi),
				B: 0.75 + 0.25*math.Cosf((phase+0.666)*math.TwoPi),
				A: 0.50 + 0.50*math.Cosf(phase*10),
			}
			phase += 0.01
			driver.Call(func() {
				label.SetColor(c)
			})
		}
	}()
	//
	
	//////////////////////
	splitter := theme.CreateSplitterLayout()
	splitter.SetOrientation(gxui.Horizontal)
	
	splitter.AddChild(directories)
	splitter.AddChild(files)


	

	topLayout := theme.CreateLinearLayout()
	topLayout.SetDirection(gxui.TopToBottom)
	//
	label1 := theme.CreateLabel()
	label1.SetColor(gxui.Blue)
	label1.SetText("Choose URL")

	//urlBox.AddChild(label1)
	//
	topLayout.AddChild(fullpath)
	topLayout.AddChild(urlBox)
	topLayout.AddChild(splitter)
	//////////////////////
	//
	
	

	btmLayout := theme.CreateLinearLayout()
	btmLayout.SetDirection(gxui.BottomToTop)
	btmLayout.SetHorizontalAlignment(gxui.AlignRight)


	btmLayout.AddChild(progressBar)
	btmLayout.AddChild(open)
	btmLayout.AddChild(downloadButton)
	btmLayout.AddChild(openFolderButton)


	btmLayout.AddChild(topLayout)
	

	window.AddChild(btmLayout)

	//////////////////
	window.OnClose(driver.Terminate)
	window.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
	//////////////////

	
	//
	//window.OnClose(ticker.Stop)
	//window.OnClose(driver.Terminate)

}

func main() {
	gl.StartDriver(appMain)
}
