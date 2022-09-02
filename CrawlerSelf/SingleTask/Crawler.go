package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"
)

//var profile []map[string]string
var infoOutCh chan map[string]string
var wg sync.WaitGroup

//var infoPassCh

const PageSite = "https://www.zhenai.com/zhenghun"
const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36"
const CityUrlMatch = `<a href="(?P<url>.*?)" data-.*?>(?P<location>.*?)</a>`
const CityDivMatch = `<dl class="city-list.*?>(?P<href>.*?)</dl>`
const PoolSize = 4

func main() {
	// 请求工具 - 用于获取网页信息
	pageRaw := crawlerBot(PageSite)
	// 数据分离 - 拿到各个城市的连接地址
	cityList := getCityList(pageRaw)

	infoPassCh := make(chan string)
	infoOutCh = make(chan map[string]string)
	// 启动进程池 -- 4 个 goroutine 去接收通道中
	// 必须先对consumer 进行初始化，不然程序会被一直堵塞住
	initPool(infoPassCh)

	for _, url := range cityList {
		infoPassCh <- url
	}
	close(infoPassCh)
	close(infoOutCh)
	//需要加一个进程同步 不然主进程会提前结束
	wg.Wait()

	// 根据链接地址去请求对应页面并拿到该城市下的第一页用户数据

	//todo: 将爬取到的信息写入一个文件中，并按照获取到的城市进行分类
	// todo: 把这个项目放到GitHub 然后形成提交记录
}

func infoOut() {
	for val := range infoOutCh {
		preety, err := json.MarshalIndent(val, "", "")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s \n", preety)
	}
	wg.Done()
}

func initPool(in chan string) {

	wg.Add(PoolSize + 1)
	for i := 0; i < PoolSize; i++ {
		go crawlerWorker(in)
	}
	go infoOut()
}

func crawlerWorker(ch chan string) {
	for url := range ch {
		pageRaw := crawlerBot(url)
		getPersonProfile(pageRaw)
	}
	wg.Done()

}

func getPersonProfile(raw []byte) {
	tableRegxp := `<tbody>.*?<tbody>`
	attrRegxp := `<td.*?><span.*?>(?P<attr>.*?)</span>(?P<value>.*?)</td>`
	nameRegxp := `<th><a.*?>(?P<nickName>.*?)</a></th>`
	tables, _ := getRegexp(tableRegxp, raw)
	for _, block := range tables {
		attr, _ := getRegexp(attrRegxp, block[0])
		nickName, _ := getRegexp(nameRegxp, block[0])
		profileIntegrate(attr, nickName)
	}
}

func profileIntegrate(attrs, NickName [][][]byte) {

	personal := make(map[string]string)
	// 属性整合
	for _, attr := range attrs {
		attrName := string(attr[1])
		attrValue := string(attr[2])
		personal[attrName] = attrValue
	}
	personal["nickName"] = string(NickName[0][1])

	infoOutCh <- personal
	//profile = append(profile, personal)
}

// crawlerBot 返回对应的链接的网页body内容
func crawlerBot(url string) []byte {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	request.Header.Add("User-Agent", UserAgent)
	reponse, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}
	// 获取网页内容
	body, err := ioutil.ReadAll(reponse.Body)
	if err != nil {
		panic(err)
	}
	return body
}

func getCityList(source []byte) map[string]string {
	cityDivMatches, _ := getRegexp(CityDivMatch, source)
	cityLinkMatches, _ := getRegexp(CityUrlMatch, cityDivMatches[0][1])
	cityListMap := make(map[string]string)

	for _, match := range cityLinkMatches {
		location, url := string(match[2]), string(match[1])
		cityListMap[location] = url
	}

	//prettyJson, _ := json.MarshalIndent(cityListMap, "", "")
	//fmt.Printf("%s", prettyJson)
	return cityListMap
}

// 没有办法再去做进一步的方法抽象  就优先返回分组的相关信息和内容
func getRegexp(regExp string, search []byte) ([][][]byte, []string) {
	if len(search) < 1 {
		return [][][]byte{}, []string{}
	}
	re := regexp.MustCompile(regExp)
	matches := re.FindAllSubmatch(search, -1)
	groupName := re.SubexpNames()

	return matches, groupName
}

//func getRegexpMap2(regExp string, search []byte) map[string][]byte {
//	if len(search) < 1 {
//		return map[string][]byte{}
//	}
//	re := regexp.MustCompile(regExp)
//	matches := re.FindAllSubmatch(search, -1)
//	groupName := re.SubexpNames()
//	res := make(map[int]map[string][]byte)
//
//	for index, match := range matches {
//		res[index]["url"] = match[1]
//		res[index]["location"] = match[2]
//	}
//	//fmt.Println(res)
//	//prettyRes, _ := json.MarshalIndent(result, "", "")
//	return res
//}
