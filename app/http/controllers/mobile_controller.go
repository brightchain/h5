package controllers

import (
	"fmt"
	"h5/pkg/model"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type MobileController struct{}

type Car_phone_model struct {
	Name   string `gorm:"column:name"`
	Brand  string `gorm:"column:brand"`
	Sort   int    `gorm:"column:sort"`
	Status int    `gorm:"column:status"`
}

// 添加表名方法
func (Car_phone_model) TableName() string {
	return "car_phone_model"
}
func (mc *MobileController) Get(c *gin.Context) {
	path := "mobile/"
	files, err := os.ReadDir(path)
	if err != nil {
		c.String(500, "Error reading directory: %v", err)
		return
	}
	db := model.RDB[model.MASTER]
	brand := "苹果"
	models := make([]Car_phone_model, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		fmt.Println(file.Name())
		if strings.Index(fileName, "黑") == -1 {
			replacer := strings.NewReplacer("_", " ", ".png", "_v.png", " ", "")
			newFileName := replacer.Replace(fileName)
			os.Rename(path+file.Name(), path+newFileName)
			// 创建白色背景图片
			originalFilePath := path + newFileName
			whiteBackgroundFilePath := path + strings.Replace(newFileName, "_v.png", "_c.png", -1)
			err = create(originalFilePath, whiteBackgroundFilePath)
			if err != nil {
				fmt.Printf("创建白色背景图片失败: %v\n", err)
			}

			models = append(models, Car_phone_model{
				Name:   strings.Replace(fileName, "_v.png", "", -1),
				Brand:  brand,
				Sort:   0,
				Status: 0,
			})
		}
	}

	if len(models) > 0 {
		fmt.Println(models)
		err := db.Db.Create(&models).Error
		if err != nil {
			c.String(500, "Error inserting data: %v", err)
			return
		}
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		if strings.Index(fileName, "黑") != -1 {
			replacer := strings.NewReplacer("黑", "", " 黑", "")
			os.Rename(path+file.Name(), path+replacer.Replace(fileName))
		}
	}
}

func create(filePath string, saveFile string) error {
	// 创建文件
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("无法解码图片: %v", err)
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 创建新的图片
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))
	// 绘制白色背景
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			newImg.Set(x, y, color.White)
		}
	}

	//保存文件
	outFile, err := os.Create(saveFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = png.Encode(outFile, newImg)
	if err != nil {
		return fmt.Errorf("无法保存图片: %v", err)
	}

	return nil
}
