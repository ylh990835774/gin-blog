package main

import (
	"log"
	"github.com/robfig/cron"
	"gin-blog/models"
	"time"
)

func main() {
	log.Println("Starting...")

	c := cron.New()
	c.AddFunc("* * * * * *", func() {
		log.Println("Run modles.CleanAllTag...")
		models.CleanAllTag()
	})
	c.AddFunc("* * * * * *", func() {
		log.Println("Run modles.CleanAllArticle...")
		models.CleanAllArticle()
	})

	c.Start()

	// time.NewTimer + for + select + t1.Reset
	// 会创建一个新的定时器，持续你设定的时间 d 后发送一个 channel 消息
	t1 := time.NewTimer(time.Second * 10)
	// for + select: 阻塞 select 等待 channel
	for {
		select {
		case <-t1.C:
			// t1.Reset: 会重置定时器，让它重新开始计时 （注意，本文适用于 “t.C已经取走，可直接使用 Reset”）
			t1.Reset(time.Second * 10)
		}
	}
}
