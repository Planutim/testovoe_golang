package main

import (
	"net/http"
	"io/ioutil"
	"path/filepath"
	"fmt"
	"encoding/json"
	"strconv"
	"image/jpeg"
	"image/png"
	"image/gif"
	"os"
	// "os/signal"
	// "syscall"
	//"log"
	// "github.com/davecgh/go-spew/spew"
	"github.com/nfnt/resize"
)
type ImageInfo struct{
	Id int `json:"id"`
	Filename string `json:"filename"`
	Url string `json:"url"`
}

var images = []ImageInfo{}

func getIndexPage(w http.ResponseWriter, r *http.Request){
	http.ServeFile(w,r,"./static/index.html")
}
func upload(w http.ResponseWriter, r *http.Request){
	r.ParseMultipartForm(10<<20)
	file,handler,err := r.FormFile("image")
	if err != nil {
		fmt.Println("Could not upload file")
		return
	}
	defer file.Close()

	tempFile, err := ioutil.TempFile("temp-images", fmt.Sprintf("*-%v", handler.Filename))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tempFile.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	tempFile.Write(bytes)
	filename := tempFile.Name()[12:]
	result := ImageInfo{
		Id: len(images)+1,
		Filename: filename,
		Url: fmt.Sprintf("http://localhost:8081/image/%v", filename),
	}
	images = append(images, result)
	json.NewEncoder(w).Encode(result)
}
func resizeImage(w http.ResponseWriter, r *http.Request){
	id := r.FormValue("id")
	ww := r.FormValue("w")
	hh :=  r.FormValue("h")
	idInt, err :=  strconv.Atoi(id)
	if err != nil{
		fmt.Println(err)
		return
	}
	if len(images)==0 || idInt>len(images){
		fmt.Println("NO IMAGES | WRONG ID")
		return
	}
	if ww == "" && hh ==""{
		fmt.Println("NO PARAMS TO RESIZE")
		return
	}
	width, err := strconv.Atoi(ww)
	if err != nil{
		fmt.Println("width absent, using 0")
	}
	height,err := strconv.Atoi(hh)
	if err != nil{
		fmt.Println("height absent, using 0")
	}
	var imgInfo ImageInfo
	
	for _, image := range images{
		if image.Id == idInt{
			imgInfo = image
			break
		}
	}

	file,err := os.Open("./temp-images/"+imgInfo.Filename)
	if err != nil{
		fmt.Println(err)
		return
	}
	defer file.Close()
	
	file_ext := filepath.Ext(imgInfo.Filename)
	fmt.Println(file_ext)
	out,err := os.Create(fmt.Sprintf("./temp-images/resized-%v",imgInfo.Filename))
	if err != nil{
		fmt.Println(err)
		return
	}
	defer out.Close()
	switch file_ext{
	case ".jpg", ".jpeg":
		image,err := jpeg.Decode(file)
		if err != nil {
			fmt.Println(err)
			return
		}
		m := resize.Resize(uint(width),uint(height),image,resize.Lanczos3)
		jpeg.Encode(out,m,nil)
		fmt.Println("JPEGGG")
	case ".png":
		image, err := png.Decode(file)
		if err != nil{
			fmt.Println(err)
			return
		}
		m := resize.Resize(uint(width),uint(height),image,resize.Lanczos3)
		png.Encode(out,m)
		fmt.Println("PNGNGGG")
	case ".gif":
		image, err := gif.Decode(file)
		if err != nil{
			fmt.Println(err)
			return
		}
		m := resize.Resize(uint(width),uint(height),image,resize.Lanczos3)
		gif.Encode(out,m,nil)
		fmt.Println("GIFFFF")
	default:
		fmt.Println("Wrong format. Cannot resize")
	}
	resized := struct{
		Filename string `json:"filename"`
		Url string `json:"url"`
	}{
		Filename: out.Name()[14:],
		Url: "http://localhost:8081/image/"+out.Name()[14:],
	}
	fmt.Printf("\nRESIZING \nid: %v\n width: %v\n height: %v\n\n", id, width, height)
	json.NewEncoder(w).Encode(resized)
}
//
func removeImages(){
	d,err := os.Open("./temp-images")
	if err != nil{
		fmt.Println(err)
		return
	}
	defer d.Close()
	names, err := d.Readdirnames(-1) //all names in a slice
	if err != nil{
		fmt.Println(err)
		return
	}
	for _, name := range names{
		err = os.RemoveAll(filepath.Join("./temp-images",name))
		if err != nil{
			fmt.Println(err)
			return
		}
	}
}
func enableServer(){
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.Handle("/image/", http.StripPrefix("/image/", http.FileServer(http.Dir("./temp-images"))))
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/resize", resizeImage)
	http.ListenAndServe(":8081", nil)
	// sigs := make(chan os.Signal, 1)
	// signal.Notify(sigs, syscall.SIGINT,syscall.SIGTERM)
	// go func(){
	// 	<- sigs
	// 	removeImages()
	// 	os.Exit(0)
	// }()
}
func main(){
	os.RemoveAll("./temp-images/")
	os.MkdirAll("./temp-images/",os.ModePerm)
	enableServer()
}