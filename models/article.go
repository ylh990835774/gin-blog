package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Article struct {
	Model

	// gorm:index，用于声明这个字段为索引，如果你使用了自动迁移功能则会有所影响，在不使用则无影响
	// Article 有一个结构体成员是 TagID，就是外键。gorm 会通过类名+ID的方式去找到这两个类之间的关联关系
	TagID int `json:"tag_id" gorm:"index"`
	// Tag字段，实际是一个嵌套的struct，它利用 TagID 与Tag模型相互关联，在执行查询的时候，能够达到Article、Tag关联查询的功能
	// 可以通过Related进行关联查
	Tag Tag `json:"tag"`

	Title      string `json:"title"`
	Desc       string `json:"desc"`
	Content    string `json:"content"`
	CreatedBy  string `json:"created_by"`
	ModifiedBy string `json:"modified_by"`
	State      int    `json:"state"`
}

func ExistArticleByID(id int) bool {
	var article Article
	db.Select("id").Where("id = ?", id).First(&article)

	if article.ID > 0 {
		return true
	}

	return false
}

func GetArticleTotal(maps interface{}) (count int) {
	db.Model(&Article{}).Where(maps).Count(&count)
	return
}

func GetArticles(pageNum int, pageSize int, maps interface{}) (articles []Article) {
	// Preload就是一个预加载器，
	// 它会执行两条SQL，分别是SELECT * FROM blog_articles;和SELECT * FROM blog_tag WHERE id IN (1,2,3,4);
	// 在查询出结构后，gorm内部处理对应的映射逻辑，将其填充到 Article 的 Tag 中, 并且避免了循环查询
	db.Preload("Tag").Where(maps).Offset(pageNum).Limit(pageSize).Find(&articles)
	return
}

func GetArticle(id int) (article Article) {
	db.Where("id = ?", id).First(&article)
	db.Model(&article).Related(&article.Tag)

	return
}

func EditArticle(id int, data interface{}) bool {
	db.Model(&Article{}).Where("id = ?", id).Update(data)
	return true
}

func AddArticle(data map[string]interface{}) bool {
	db.Create(&Article{
		TagID:     data["tag_id"].(int),
		Title:     data["title"].(string),
		Desc:      data["desc"].(string),
		Content:   data["content"].(string),
		CreatedBy: data["created_by"].(string),
		State:     data["state"].(int),
	})
	return true
}

func DeleteArticle(id int) bool {
	db.Where("id = ?", id).Delete(Article{})
	return true
}

func (article *Article) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedOn", time.Now().Unix())
	return nil
}

func (article *Article) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("ModifiedOn", time.Now().Unix())
	return nil
}

// 硬删除: deleted_on 不为空的数据, 配合定时器(cron.go)使用
func CleanAllArticle() bool {
	db.Unscoped().Where("deleted_on != ?", 0).Delete(&Article{})
	return true
}
