package controllers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)


type video struct {
	Video *multipart.FileHeader `form:"video" binding:"required"`
	Fw string 	`form:"fw" binding:"required"`		// frame width in px
	Fh string 	`form:"fh" binding:"required"`		// frame height in px
	X string 	`form:"x" binding:"required"`		//  x position of video top left
	Y string 	`form:"y" binding:"required"`		// y position of video top left
	Vw string 	`form:"vw" binding:"required"`		// video width
	Vh string 	`form:"vh" binding:"required"`		// video height
}

type videoMeta struct {
	Fw int
	Fh int
	X int
	Y int
	Vw int
	Vh int
}

func Get(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "okay",
	})
}

func cropFromTop(y int, height int, width int) string {
	return fmt.Sprintf("crop=%d:%d:%d:%d", width, height - y, 0, y);
}

func cropFromBottom(yB int, height int, width int) string {
	return fmt.Sprintf("crop=%d:%d:%d:%d", width, height - yB, 0, 0);
}

func cropFromLeft(x int, height int, width int) string {
	return fmt.Sprintf("crop=%d:%d:%d:%d", width - x, height, x, 0);
}

func cropFromRight(xR int, height int, width int) string {
	return fmt.Sprintf("crop=%d:%d:%d:%d", width - xR, height, 0, 0);
}

func getScaleFilterValue(vmd videoMeta) string {
	filterArray := []string{
		fmt.Sprintf("scale=%d:%d", vmd.Vw, vmd.Vh),
		"pad='width=ceil(iw/2)*2:height=ceil(ih/2)*2'",
	}

	if (vmd.Y < 0) {
		filterArray = append(
			filterArray, 
			cropFromTop(-1 * vmd.Y, vmd.Vh, vmd.Vw),
		);
		vmd.Vh = vmd.Vh + vmd.Y;
		vmd.Y = 0;
	}

	if (vmd.X < 0) {
		filterArray = append(
			filterArray, 
			cropFromLeft(-1 * vmd.X, vmd.Vh, vmd.Vw),
		);
		vmd.Vw = vmd.Vw + vmd.X;
		vmd.X = 0;
	}

	if ((vmd.X + vmd.Vw) > vmd.Fw) {
		rightDiff := (vmd.X + vmd.Vw) - vmd.Fw;
		filterArray = append(
			filterArray, 
			cropFromRight(rightDiff, vmd.Vh, vmd.Vw),
		);
		vmd.Vw = vmd.Vw - rightDiff;
	}
	
	if ((vmd.Y + vmd.Vh) > vmd.Fh) {
		bottomDiff := (vmd.Y + vmd.Vh) - vmd.Fh;
		filterArray = append(
			filterArray, 
			cropFromBottom(bottomDiff, vmd.Vh, vmd.Vw),
		);
		vmd.Vh = vmd.Vh - bottomDiff;
	}

	filterArray = append(
		filterArray, 
		fmt.Sprintf("pad=%d:%d:%d:%d:black", vmd.Fw, vmd.Fh, vmd.X, vmd.Y),
	);

	return strings.Join(filterArray[:], ",");
}

func transcodeVideo(inputFilename string, outputFilename string, filterString string) {
	err := ffmpeg_go.Input(inputFilename).
	Output(outputFilename,ffmpeg_go.KwArgs{
		"vcodec": "libx264",
		"r": 24,
		"movflags": "+faststart",
		"pix_fmt": "yuv420p",
		"preset": "fast",
		"vf":filterString,
	}).
	OverWriteOutput().
	Run()

	if err != nil {
		fmt.Println("Error Ocurred while transcoding")
	}
}

// we are building for devices that have ratio 9x16 (atleast: 720x1280)
func PutUser(c *gin.Context) {
	var videoObj video
	if err := c.ShouldBind(&videoObj); err != nil {
		c.String(http.StatusBadRequest, "bad request 1")
		return
	}

	tempFilePath := "assets/input.mov";
	inputVideo := videoObj.Video;
	if err := c.SaveUploadedFile(inputVideo, tempFilePath); err != nil {
		c.String(http.StatusInternalServerError, "could not upload input video")
		return
	}

	fw,err:= strconv.Atoi(videoObj.Fw)
	fh,err:= strconv.Atoi(videoObj.Fh)
	x,err:= strconv.Atoi(videoObj.X)
	y,err:= strconv.Atoi(videoObj.Y)
	vw,err:= strconv.Atoi(videoObj.Vw)
	vh,err:= strconv.Atoi(videoObj.Vh)
	if err != nil {
		c.String(http.StatusBadRequest, "bad request 2")
		return
	}

	videoMetaData := videoMeta{
		fw,
		fh,
		x,
		y,
		vw,
		vh,
	}

	filterString := getScaleFilterValue(videoMetaData);
	fmt.Println(filterString);

	transcodeVideo(tempFilePath, "assets/output.mp4",filterString);

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": videoObj,
	})
}