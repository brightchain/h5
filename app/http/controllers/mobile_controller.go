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
	brand := "红米"
	models := make([]Car_phone_model, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		fmt.Println(file.Name())
		if strings.Index(fileName, "黑") == -1 {
			replacer := strings.NewReplacer("_", " ", ".png", "_v.png", " ", "",)
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
				Name:   strings.Replace(newFileName, "_v.png", "", -1),
				Brand:  brand,
				Sort:   0,
				Status: 1,
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
			replacer := strings.NewReplacer("黑", "", " 黑", "","_", " "," ", "")
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

// 转换PNG图片：透明->白色，白色->透明
func convertPNGTransparency(inputPath, outputPath string) error {
	// 打开原图片文件
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("无法打开文件: %v", err)
	}
	defer file.Close()

	// 解码PNG图片
	img, err := png.Decode(file)
	if err != nil {
		return fmt.Errorf("解码PNG失败: %v", err)
	}

	// 获取图片尺寸
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// 创建新的RGBA图片
	newImg := image.NewRGBA(bounds)

	// 遍历每个像素进行转换
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// 获取原像素颜色
			originalColor := img.At(x, y)
			r, g, b, a := originalColor.RGBA()

			// 转换为0-255范围
			rByte := uint8(r >> 8)
			gByte := uint8(g >> 8)
			bByte := uint8(b >> 8)
			aByte := uint8(a >> 8)

			var newColor color.RGBA

			// 判断是否为透明像素（alpha < 128）
			if aByte < 128 {
				// 透明部分转为白色
				newColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
			} else {
				// 判断是否为白色或接近白色
				if isWhiteish(rByte, gByte, bByte) {
					// 白色部分转为透明
					newColor = color.RGBA{R: 0, G: 0, B: 0, A: 0}
				} else {
					// 其他颜色保持不变
					newColor = color.RGBA{R: rByte, G: gByte, B: bByte, A: aByte}
				}
			}

			newImg.Set(x, y, newColor)
		}
	}

	// 创建输出文件
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outFile.Close()

	// 编码并保存PNG
	err = png.Encode(outFile, newImg)
	if err != nil {
		return fmt.Errorf("编码PNG失败: %v", err)
	}

	return nil
}

// 判断是否为白色或接近白色
func isWhiteish(r, g, b uint8) bool {
	// 可以调整这个阈值来控制什么算作"白色"
	threshold := uint8(240)
	return r >= threshold && g >= threshold && b >= threshold
}

// 更精确的白色检测（考虑色彩偏差）
func isWhiteishAdvanced(r, g, b uint8) bool {
	// 计算RGB的平均值
	avg := (int(r) + int(g) + int(b)) / 3
	
	// 检查是否接近白色且RGB值相近
	if avg >= 240 {
		maxDiff := 15 // 允许的最大色差
		return abs(int(r)-avg) <= maxDiff && 
		       abs(int(g)-avg) <= maxDiff && 
		       abs(int(b)-avg) <= maxDiff
	}
	return false
}

// 绝对值函数
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// 批量处理版本
func batchConvertPNG(inputDir, outputDir string) error {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %v", err)
	}

	// 确保输出目录存在
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		// 检查是否为PNG文件
		if len(filename) > 4 && filename[len(filename)-4:] == ".png" {
			inputPath := inputDir + "/" + filename
			outputPath := outputDir + "/" + filename
			
			fmt.Printf("处理文件: %s\n", filename)
			err := convertPNGTransparency(inputPath, outputPath)
			if err != nil {
				fmt.Printf("处理 %s 失败: %v\n", filename, err)
			} else {
				fmt.Printf("完成: %s\n", filename)
			}
		}
	}
	return nil
}

// 带配置选项的高级转换函数
type ConvertOptions struct {
	WhiteThreshold   uint8 // 白色判断阈值 (0-255)
	TransparentThreshold uint8 // 透明判断阈值 (0-255)
	ColorTolerance   uint8 // 颜色容差
}

func convertPNGWithOptions(inputPath, outputPath string, opts ConvertOptions) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return err
	}

	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			r, g, b, a := originalColor.RGBA()

			rByte := uint8(r >> 8)
			gByte := uint8(g >> 8)
			bByte := uint8(b >> 8)
			aByte := uint8(a >> 8)

			var newColor color.RGBA

			// 判断是否为透明
			if aByte < opts.TransparentThreshold {
				// 透明->白色
				newColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
			} else if isWhiteWithTolerance(rByte, gByte, bByte, opts.WhiteThreshold, opts.ColorTolerance) {
				// 白色->透明
				newColor = color.RGBA{R: 0, G: 0, B: 0, A: 0}
			} else {
				// 保持原色
				newColor = color.RGBA{R: rByte, G: gByte, B: bByte, A: aByte}
			}

			newImg.Set(x, y, newColor)
		}
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return png.Encode(outFile, newImg)
}

func isWhiteWithTolerance(r, g, b, threshold, tolerance uint8) bool {
	return r >= threshold && g >= threshold && b >= threshold &&
		   abs(int(r)-int(g)) <= int(tolerance) &&
		   abs(int(r)-int(b)) <= int(tolerance) &&
		   abs(int(g)-int(b)) <= int(tolerance)
}

func main() {
	// 示例1: 单个文件转换
	fmt.Println("开始转换PNG透明度...")
	
	err := convertPNGTransparency("input.png", "output.png")
	if err != nil {
		fmt.Printf("转换失败: %v\n", err)
		return
	}
	fmt.Println("转换完成!")

	// 示例2: 批量转换
	// err = batchConvertPNG("./input", "./output")
	// if err != nil {
	//     fmt.Printf("批量转换失败: %v\n", err)
	//     return
	// }

	// 示例3: 使用自定义选项
	// opts := ConvertOptions{
	//     WhiteThreshold:       240,
	//     TransparentThreshold: 128,
	//     ColorTolerance:       15,
	// }
	// err = convertPNGWithOptions("input.png", "output_custom.png", opts)
	// if err != nil {
	//     fmt.Printf("自定义转换失败: %v\n", err)
	// }
}
