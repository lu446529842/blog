package controllers

import (
	"blog/app/models"
	"strconv"
	"strings"
	"time"

	"github.com/revel/revel"
)

/**
 * Add a blog for admin user
 * 发布博客 action
 */

var blogModel *models.Blogger

// PostData model.
// 发布博客前端提交的数据
type PostData struct {
	Id          int64
	Title       string //博客标题
	ContentMD   string //博客内容 MD
	ContentHTML string // 博客内容 HTML
	Category    int64  // 博客类别
	Tag         string // 标签 格式：12,14,32
	Keywords    string // 关键词 格式：java,web开发
	passwd      string //博客内容是否加密
	Summary     string // 博客摘要
	Type        int    // 0 表示 markdown，1 表示 html
	NewTag      string // 新添加的标签
	Createtime  string //创建时间
}

// User for User Controller
type Post struct {
	Admin
}

// 创建博客页面
func (p *Post) Index(postid int64) revel.Result {
	categoryModel := new(models.Category)
	tagModel := new(models.BloggerTag)
	p.RenderArgs["categorys"] = categoryModel.FindAll()
	tags, err := tagModel.ListAll()
	if err != nil {
		tags = make([]models.BloggerTag, 0)
	}
	blog := &models.Blogger{Id: postid}
	if postid > 0 {
		blog, err = blog.FindById()
		if err != nil {
			p.NotFound("博客不存在")
		} else {
			p.RenderArgs["blog"] = blog
		}
	}
	p.RenderArgs["tags"] = tags
	p.RenderArgs["timenow"] = time.Now()
	return p.RenderTemplate("Admin/Post/Index.html")
}

//ManagePost .
// 管理博客页面
func (p *Post) ManagePost(uid, category int64) revel.Result {
	blogs, err := blogModel.GetBlogByPageAND(uid, category, 1, 20)
	if err != nil {
		blogs = make([]models.Blogger, 0)
	}
	p.RenderArgs["blogs"] = blogs
	p.RenderArgs["p_uid"] = uid
	p.RenderArgs["p_ca"] = category
	return p.RenderTemplate("Admin/Post/Manage-post.html")
}

// NewPostHandler to Add new article.
// 添加博客
func (p *Post) NewPostHandler() revel.Result {
	data := new(PostData)
	p.Params.Bind(&data, "data")
	p.Validation.Required(data.Title).Message("标题不能为空")
	p.Validation.Required(data.ContentHTML).Message("内容不能为空")

	if p.Validation.HasErrors() {
		return p.RenderJson(&ResultJson{Success: false, Msg: p.Validation.Errors[0].Message})
	}

	blog := new(models.Blogger)
	blog.Title = data.Title
	blog.ContentHTML = data.ContentHTML
	blog.ContentMD = data.ContentMD
	blog.CategoryId = data.Category
	blog.Type = data.Type
	blog.Summary = data.Summary

	// 处理创建时间
	tm, err := time.Parse("2006-01-02", data.Createtime)
	if err != nil {
		blog.CreateTime = time.Now()
	} else {
		blog.CreateTime = tm
	}

	uid := p.Session["UID"]
	authorid, _ := strconv.Atoi(uid)
	blog.CreateBy = int64(authorid)

	if data.passwd != "" {
		blog.Passwd = data.passwd
	}

	var blogID int64
	if data.Id > 0 {
		blog.Id = data.Id
		_, err = blog.Update()
		if err != nil {
			revel.ERROR.Println("博客更新失败：", err)
		}
		blogID = data.Id
	} else {
		blogID, err = blog.New()
	}

	// 添加新的标签
	btr := new(models.BloggerTagRef)
	newTags := strings.Split(data.NewTag, ",")
	for _, v := range newTags {
		tag := &models.BloggerTag{Name: v}
		tagid, _ := tag.New()
		if tagid > 0 {
			btr.AddTagRef(tagid, blogID)
		}
	}

	// 处理标签关联
	blog.DeleteAllBlogTags()
	tagids := strings.Split(data.Tag, ",")
	for _, v := range tagids {
		id, err := strconv.Atoi(v)
		if err == nil {
			btr.AddTagRef(int64(id), blogID)
		}
	}

	if err != nil || blogID <= 0 {
		p.Flash.Error("msg", "create new blogger post error.")
		return p.RenderJson(&ResultJson{Success: false, Msg: err.Error(), Data: ""})
	}
	return p.RenderJson(&ResultJson{Success: true})
}

func (p *Post) QueryCategorys() revel.Result {
	c := new(models.Category)
	arr := c.FindAll()
	return p.RenderJson(&ResultJson{Success: true, Msg: "", Data: arr})
}

func (p *Post) CreateTag(name string) revel.Result {
	tag := new(models.BloggerTag)
	tag.Name = name
	tag.Parent = 0
	tag.Type = 0
	_, err := tag.New()
	if err != nil {
		return p.RenderJson(&ResultJson{Success: false, Msg: err.Error(), Data: ""})
	}
	return p.RenderJson(&ResultJson{Success: true, Msg: "", Data: ""})
}

func (p *Post) Delete(ids string) revel.Result {
	idArr := strings.Split(ids, ",")
	if len(idArr) > 0 {
		for _, v := range idArr {
			id, err := strconv.Atoi(v)
			if err == nil {
				blog := &models.Blogger{Id: int64(id)}
				blog.Del()
			}
		}
		return p.RenderJson(&ResultJson{Success: true})
	} else {
		return p.RenderJson(&ResultJson{Success: true})
	}
}
