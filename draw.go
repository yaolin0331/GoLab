
/**
2020 Jill in clone 寫好玩的Go抽籤代碼
讀取 json 抽出後，重寫 json
**/

package main

import (
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

type Lists struct {
	Name            string
	HostData        string
}

var AvailableList []*Lists

func main() {
	viper.SetConfigName("db") // 设置配置文件名 (不带后缀)
	viper.AddConfigPath(".")      // 第一个搜索路径
	err := viper.ReadInConfig()   // 读取配置数据
	if err != nil {
		panic(fmt.Errorf("Fatal error db.json file: %s \n", err))
	}
	filter()

}

func filter() {
	//間隔N天
	interval := 3
	//主持日期(今天或最近一天工作日週一到週五，風雨無阻)
	hostDate := getNextworkingDay()
	//篩選，N天內沒有住持者名單
	//在{}日期內
	checkDates := getNWorkingDates(interval, hostDate)
	fmt.Println("已主持過的日期:::", checkDates)

	AvailableList = make([]*Lists, 0)
	all := viper.GetStringMapString("people")
	for k, v := range all {
		if IsExist(v, checkDates) {
			continue
		}
		list := new(Lists)
		list.Name = k
		list.HostData = v
		AvailableList = append(AvailableList, list)
		fmt.Println(k, v)
	}

	fmt.Println("本次入選主持人數==>", len(AvailableList))

	draw(all, hostDate)
}

/**
	篩選 N 日內沒有被抽出的名單
 */
func draw(all map[string]string, hostDate string) {

	// draw a luck host
	r:= generateRangeNum(0, len(AvailableList))
	fmt.Println(AvailableList[r])

	fmt.Println("**************")
	fmt.Println("恭喜下次主持人=>", AvailableList[r].Name, "，上次主持日期時間是=>", AvailableList[r].HostData)
	fmt.Println("**************")

	fmt.Println("同意 (y/n) ?  不同意，幾秒後將自動重新抽籤 ! ")

	// prompt  if OK
	// 回寫回 db.json 否則 重抽，抽過的人標注hostdate，可躲過N天后在重抽

	if askForConfirmation() {
		AvailableList[r].HostData = hostDate
		reWriteDB(all, AvailableList[r])

	} else {
		GotoSleep()
		fmt.Println("---- 將進行重次抽取 ----")
		draw(all, hostDate)

	}

}


/**
暫停隨機數(0~10)秒
*/
func GotoSleep() {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(6) // n will be between 0 and 100
	fmt.Printf("Sleeping %d seconds...\n", n)
	time.Sleep(time.Duration(n) * time.Second)
}

func askForConfirmation() bool {
	var response string

	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println(err)
		return false
	}

	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		fmt.Println("您輸入的無法辨識，請輸入 (y)es 或 (n)o 後再按 enter！")
		return askForConfirmation()
	}
}

func reWriteDB(all map[string]string, luckhost *Lists) {

	content := `{` + "\n" + `"people": {` + "\n"
	c := 0
	for k, v:=range all {
		// 前已經抽過，須重抽
		if v==luckhost.HostData {
			v=""
		}
		//fmt.Println(k, v)
		if k==luckhost.Name {
			v=luckhost.HostData
		}
		content += "\t" + `"` + strings.Title(strings.ToLower(k)) + `":"` + v + `"`
		c++
		if c < len(all) {
			content += `,` + "\n"
		}
	}
	content += "\n" + `}`+ "\n" +`}`

	fmt.Println("更新db=>", content)

	//將指定內容寫入到檔案中
	err := ioutil.WriteFile("db.json", []byte(content), 0666)
	if err != nil {
		fmt.Println("ioutil WriteFile error: ", err)
	}

}

// GenerateRangeNum 生成一個區間範圍的隨機數
func generateRangeNum(min, max int) int {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max - min) + min
	fmt.Printf("rand is %v\n", randNum)
	return randNum
}

func getNextworkingDay() string {

	var dt = time.Now()
	drawDate := dt.Format("20060102")
	fmt.Println("抽籤日期=>", drawDate)

	for i:=1; i<30; i++ {
		dt = time.Now().AddDate(0, 0, i)
		if IsExist(dt.Weekday().String(), []string{"Saturday","Sunday"}) {
			continue
		}
		break
	}
	hostDate := dt.Format("20060102")
	fmt.Println("下次主持日期=>", hostDate, dt.Weekday().String())

	return hostDate

}

// 取得N個工作天來比對
func getNWorkingDates(interval int, hostDate string) []string {
	result := make([]string, interval)
	ht, _ := time.Parse("20060102", hostDate)
	idx := 0
	for i:=1; i<=interval; i++ {
		dt := ht.AddDate(0, 0, i*-1)
		if IsExist(dt.Weekday().String(), []string{"Saturday","Sunday"}) {
			interval++
			continue
		}
		result[idx] = dt.Format("20060102")
		idx++
	}
	return result
}

func IsExist(val interface{}, in interface{}) bool {
	arr := reflect.ValueOf(in)
	switch arr.Kind(){
	case reflect.Slice, reflect.Array:
		for i := 0; i < arr.Len(); i++ {
			if arr.Index(i).Interface() == val {
				return true
			}
		}

	case reflect.Map:
		if arr.MapIndex(reflect.ValueOf(val)).IsValid() {
			return true
		}

	default:
		panic("Invalid data-type")

	}
	return false

}