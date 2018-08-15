package main

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"fmt"
	"github.com/jinzhu/gorm"
	_"github.com/jinzhu/gorm/dialects/mysql"
	"github.com/qiniu/api.v7/storage"
	"github.com/qiniu/api.v7/auth/qbox"
	"strings"
	"qiniupkg.com/x/log.v7"
	"football/work"
	"sync"
	"strconv"
)

type Urls struct {
	url string
	TeamId int
}
type Team struct {
	Team_id int
	Team_name string
	Team_logo string
}
type Player struct {
	P_position string
	P_name string
	P_icon string
	P_number string
	Team_id int
}
type PutRet struct {
	Key    string
	Hash   string
	Name   string
}
var db *gorm.DB
func init()  {
	var err error
	db,err=gorm.Open("mysql","root:12345@(127.0.0.1:3306)/football?charset=utf8&parseTime=True&loc=Local")
	if err!=nil{
		panic(err)
	}
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

}
func (u *Urls)Task()  {
	getTeamPlayer(u.url,u.TeamId)
}
func downImage(url string)  string{
	domain:="http://pbjk21xhc.bkt.clouddn.com/"
	bucket:="football"
	accessKey:="*******"
	secretKey:="*******"
	mac:=qbox.NewMac(accessKey,secretKey)
	cfg:=storage.Config{}
	cfg.Zone = &storage.ZoneHuabei
	cfg.UseHTTPS = false
	cfg.UseCdnDomains = false
	bucketManager:=storage.NewBucketManager(mac,&cfg)
	fetchRet, err:=bucketManager.FetchWithoutKey(url,bucket)
	if err!=nil{
		//fmt.Println("fetch error,", err)
		return domain
	}else{
		return domain+fetchRet.Key
		//fmt.Println(fetchRet.String())
	}
}
func main()  {
	p:=work.New(10)
	var wg sync.WaitGroup
	url:="https://www.dongqiudi.com/data?competition=8"
	response,err:=http.Get(url)
	if err!=nil{
		log.Println(err)
	}
	defer response.Body.Close()
	if response.StatusCode==200{
		fmt.Println("请求成功")
	}
	doc,err:=goquery.NewDocumentFromReader(response.Body)
	if err!=nil{
		panic(err)
	}
	log.Println("开始爬取")
	doc.Find(".list_1").Find("tbody").Find(".team").Each(func(i int, selection *goquery.Selection) {
		wg.Add(100)
		team:=new(Team)
		team.Team_id=i+1

		//球队图标
		logo:=selection.Find("img").AttrOr("src","")
		//球队名字
		team.Team_name=strings.TrimSpace(selection.Find("a").Text())
		if team.Team_name!=""{
			p_url:=selection.Find("a").AttrOr("href","")
			url:=Urls{url:p_url,TeamId:team.Team_id}
			//getTeamPlayer(p_url,team.Team_id)
			go func() {
				p.Run(&url)
				wg.Done()
			}()
			team.Team_logo=downImage(logo)
			if err := db.Table("team").Create(team).Error; err != nil {
				log.Println(err)
			}
			fmt.Println(team)
		}
	})
	wg.Wait()
	p.ShutDown()
}

func getTeamPlayer(url string,team_id int)  {
	//url:="https://www.dongqiudi.com/team/50000564.html"
	response,err:=http.Get(url)
	if err!=nil{
		panic(err)
	}
	doc,err:=goquery.NewDocumentFromReader(response.Body)
	doc.Find(".teammates_list").Find("tbody").Find(".stat_list").Each(func(i int, selection *goquery.Selection) {
		//v,is:=selection.Attr()
		_,is:=selection.Attr("style")
		if is{
			//
			p:=new(Player)
			p.Team_id = team_id
			selection.Find("td").Each(func(j int, s2 *goquery.Selection) {
				//fmt.Println(j)
				if j==0{
					p.P_position = s2.Text()
				}
				if j==1{
					p.P_number = s2.Text()
				}
				if j==2{
					p.P_name = strings.TrimSpace(s2.Find("a").Text())
					icon:= s2.Find("img").AttrOr("src","")
					p.P_icon = downImage(icon)
				}
			})
			//fmt.Println(p)
			//插入到数据库
			if err := db.Table("players").Create(p).Error; err != nil {
				log.Println(err)
			}
		}
	})
	fmt.Println("爬取成功！队伍id"+ strconv.Itoa(team_id))
}


