package main

/*
#include<windows.h>
#include<versionhelpers.h>
*/
import "C"
import (
	"encoding/xml"
	"image/bmp"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"syscall"
	"unicode/utf16"
)
type Image struct{
	Date string `xml:"enddate"`
	Link string `xml:"url"`
}
type Images struct{
	Images []Image `xml:"image"`
}
func getImageInfo() *Image {
	resp,err:=http.Get("https://cn.bing.com/HPImageArchive.aspx?idx=0&n=1")
	if err!=nil {
		return nil
	}
	defer func() {
		_=resp.Body.Close()
	}()
	body,err:=ioutil.ReadAll(resp.Body)
	if err!=nil {
		return nil
	}
	imgs:=Images{}
	if _=xml.Unmarshal(body,&imgs);len(imgs.Images)!=1 {
		return nil
	}
	return &imgs.Images[0]
}
func downloadFile(dirs *string,info *Image) *string {
	path:=*dirs+"Bing"+info.Date+".jpg"
	_,err:=os.Stat(path)
	if err==nil {
		return nil
	}
	resp,err:=http.Get("https://cn.bing.com"+info.Link)
	if err!=nil {
		return nil
	}
	defer func() {
		_=resp.Body.Close()
	}()
	file,err:=os.Create(path)
	if err!=nil {
		return nil
	}
	defer func() {
		_=file.Close()
	}()
	if _,err=io.Copy(file,resp.Body);err!=nil {
		return nil
	}
	return &path
}
func jpegToBmp(dirs *string,path *string) bool {
	src,err:=os.Open(*path)
	if err!=nil {
		return false
	}
	defer func() {
		_=src.Close()
	}()
	img,err:=jpeg.Decode(src)
	if err!=nil {
		return false
	}
	*path=*dirs+".tmp"
	_=os.Remove(*path)
	dst,err:=os.Create(*path)
	if err!=nil {
		return false
	}
	defer func() {
		_=dst.Close()
	}()
	if bmp.Encode(dst,img)!=nil {
		return false
	}
	C.SetFileAttributesW((*C.ushort)(wcs(path)),2)
	return true
}
func wcs(str *string) *uint16 {
	wcs:=append(utf16.Encode([]rune(*str)),uint16(0))
	return &wcs[0]
}
func main() {
	dirs,err:=syscall.FullPath("Wallpaper/")
	if err!=nil||os.MkdirAll(dirs,os.ModePerm)!=nil {
		return
	}
	info:=getImageInfo()
	if info==nil {
		return
	}
	path:=downloadFile(&dirs,info)
	if path==nil {
		return
	}
	if C.IsWindows8OrGreater()==0&&!jpegToBmp(&dirs,path) {
		return
	}
	C.SystemParametersInfoW(20,1,C.PVOID(wcs(path)),1)
}