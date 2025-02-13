package api

import (
	"github.com/Ptt-official-app/go-openbbsmiddleware/apitypes"
	"github.com/Ptt-official-app/go-openbbsmiddleware/dbcs"
	"github.com/Ptt-official-app/go-openbbsmiddleware/queue"
	"github.com/Ptt-official-app/go-openbbsmiddleware/schema"
	"github.com/Ptt-official-app/go-openbbsmiddleware/types"
	"github.com/Ptt-official-app/go-openbbsmiddleware/utils"
	pttbbsapi "github.com/Ptt-official-app/go-pttbbs/api"
	"github.com/Ptt-official-app/go-pttbbs/bbs"
	"github.com/Ptt-official-app/go-pttbbs/ptttype"
	pttbbstypes "github.com/Ptt-official-app/go-pttbbs/types"
	"github.com/gin-gonic/gin"
)

const GET_ARTICLE_R = "/board/:bid/article/:aid"

type GetArticleDetailParams struct{}

type GetArticleDetailPath struct {
	FBoardID   apitypes.FBoardID   `uri:"bid"`
	FArticleID apitypes.FArticleID `uri:"aid"`
}

type GetArticleDetailResult struct {
	BBoardID   apitypes.FBoardID   `json:"bid"`         //
	ArticleID  apitypes.FArticleID `json:"aid"`         //
	IsDeleted  bool                `json:"deleted"`     //
	CreateTime types.Time8         `json:"create_time"` //
	MTime      types.Time8         `json:"modified"`    //
	Recommend  int                 `json:"recommend"`   //
	NComments  int                 `json:"n_comments"`  //
	Owner      bbs.UUserID         `json:"owner"`       //
	Title      string              `json:"title"`       //
	Money      int                 `json:"money"`       //
	Class      string              `json:"class"`       // can be: R: 轉, [class]
	Filemode   ptttype.FileMode    `json:"mode"`        //

	URL  string `json:"url"`  //
	Read bool   `json:"read"` //

	Brdname string          `json:"brdname"`
	Content [][]*types.Rune `json:"content"`
	IP      string          `json:"ip"`
	Host    string          `json:"host"` // ip 的中文呈現, 外國則為國家.
	BBS     string          `json:"bbs"`

	Rank int `json:"rank"` // 評價

}

func GetArticleDetailWrapper(c *gin.Context) {
	params := &GetArticleDetailParams{}
	path := &GetArticleDetailPath{}
	LoginRequiredPathQuery(GetArticleDetail, params, path, c)
}

func GetArticleDetail(remoteAddr string, userID bbs.UUserID, params interface{}, path interface{}, c *gin.Context) (result interface{}, statusCode int, err error) {
	thePath, ok := path.(*GetArticleDetailPath)
	if !ok {
		return nil, 400, ErrInvalidPath
	}

	boardID, err := toBoardID(thePath.FBoardID, remoteAddr, userID, c)
	if err != nil {
		return nil, 500, err
	}
	articleID := thePath.FArticleID.ToArticleID()

	// validate user
	_, statusCode, err = isBoardValidUser(boardID, c)
	if err != nil {
		return nil, statusCode, err
	}

	// ensure that we do have the article.
	content, ip, host, bbs, articleDetailSummary, statusCode, err := tryGetArticleContentInfo(userID, boardID, articleID, c, false)
	if err != nil {
		return nil, statusCode, err
	}

	url := apitypes.ToURL(thePath.FBoardID, thePath.FArticleID)

	result = &GetArticleDetailResult{
		BBoardID:   apitypes.ToFBoardID(articleDetailSummary.BBoardID),
		ArticleID:  apitypes.ToFArticleID(articleDetailSummary.ArticleID),
		CreateTime: articleDetailSummary.CreateTime.ToTime8(),
		MTime:      articleDetailSummary.MTime.ToTime8(),
		Recommend:  articleDetailSummary.Recommend,
		NComments:  articleDetailSummary.NComments,
		Owner:      articleDetailSummary.Owner,
		Title:      articleDetailSummary.Title,
		Money:      articleDetailSummary.Money,
		Class:      articleDetailSummary.Class,
		Filemode:   articleDetailSummary.Filemode,
		Rank:       articleDetailSummary.Rank,

		URL:  url,
		Read: true,

		Brdname: boardID.ToBrdname(),

		Content: content,
		IP:      ip,
		Host:    host,
		BBS:     bbs,
	}

	return result, 200, nil
}

//tryGetArticleContentInfo
//
//嘗試拿到 article-content
//
//1. 根據 article-id 得到相對應的 filename, ownerid, create-time.
//2. 嘗試從 schema 拿到 db summary 資訊. (create-time)
//3. 如果可以從 schema 拿到 db 資訊:
//   3.1. 如果已經 deleted: return deleted.
//   3.2. 如果距離上次跟 pttbbs 拿的時間太近: 從 schema 拿到 content, return schema-content.
//4. 嘗試做 lock.
//   4.1. 如果 lock 失敗: 從 schema 拿到 content, return schema-content.
//5. 從 pttbbs 拿到 article
//6. 如果從 pttbbs 拿到的時間比 schema 裡拿到的時間舊的話: return schema-content.
//7. parse article 為 content / comments.
//8. 將 comments parse 為 firstComments / theRestComments.
//9. 將 theRestComments 丟進 queue 裡.
func tryGetArticleContentInfo(userID bbs.UUserID, bboardID bbs.BBoardID, articleID bbs.ArticleID, c *gin.Context, isSystem bool) (content [][]*types.Rune, ip string, host string, bbs string, articleDetailSummary *schema.ArticleDetailSummary, statusCode int, err error) {
	updateNanoTS := types.NanoTS(0)

	// set user-read-article-id
	defer func() {
		if err == nil {
			setUserReadArticle(content, userID, articleID, updateNanoTS)
		}
	}()

	// get article-info (ensuring valid article-id)
	articleFilename := articleID.ToRaw()
	articleCreateTime, err := articleFilename.CreateTime()
	if err != nil {
		return nil, "", "", "", nil, 500, err
	}

	articleCreateTimeNanoTS := types.Time4ToNanoTS(articleCreateTime)

	// get from backend with content-mtime
	// estimated max 500ms + 3 seconds

	articleDetailSummary_db, statusCode, err := TryGetArticleDetailSummary(userID, bboardID, articleID, articleCreateTime, c, isSystem)
	if err != nil {
		return nil, "", "", "", nil, statusCode, err
	}

	// preliminarily checking article-detail-summary.
	if articleDetailSummary_db.IsDeleted {
		return nil, "", "", "", nil, 500, ErrAlreadyDeleted
	}

	if articleDetailSummary_db.CreateTime == 0 {
		articleDetailSummary_db.CreateTime = articleCreateTimeNanoTS
	}

	if TryGetArticleContentInfoTooSoon(articleDetailSummary_db.ContentUpdateNanoTS) {

		contentInfo, err := schema.GetArticleContentInfo(bboardID, articleID)
		if err != nil {
			return nil, "", "", "", nil, 500, err
		}
		return contentInfo.Content, contentInfo.IP, contentInfo.Host, contentInfo.BBS, articleDetailSummary_db, 200, nil
	}

	ownerID := articleDetailSummary_db.Owner

	// 4. do lock. if failed, return the data in db.
	lockKey := ArticleLockKey(bboardID, articleID)
	err = schema.TryLock(lockKey, ARTICLE_LOCK_TS_DURATION)
	if err != nil {
		contentInfo, err := schema.GetArticleContentInfo(bboardID, articleID)
		if err != nil {
			return nil, "", "", "", nil, 500, err
		}
		updateNanoTS = types.NowNanoTS()
		return contentInfo.Content, contentInfo.IP, contentInfo.Host, contentInfo.BBS, articleDetailSummary_db, 200, nil
	}
	defer func() { _ = schema.Unlock(lockKey) }()

	// 5. get article from pttbbs
	theParams_b := &pttbbsapi.GetArticleParams{
		RetrieveTS: articleDetailSummary_db.ContentMTime.ToTime4(),
	}
	var result_b *pttbbsapi.GetArticleResult

	urlMap := map[string]string{
		"bid": string(bboardID),
		"aid": string(articleID),
	}
	url := utils.MergeURL(urlMap, pttbbsapi.GET_ARTICLE_R)

	statusCode, err = utils.BackendGet(c, url, theParams_b, nil, &result_b)
	if err != nil {
		return nil, "", "", "", nil, statusCode, err
	}

	// 6. check content-mtime (no modify from backend, no need to parse again)
	contentMTime := types.Time4ToNanoTS(result_b.MTime)
	if articleDetailSummary_db.ContentMTime >= contentMTime {
		contentInfo, err := schema.GetArticleContentInfo(bboardID, articleID)
		if err != nil {
			return nil, "", "", "", nil, 500, err
		}
		return contentInfo.Content, contentInfo.IP, contentInfo.Host, contentInfo.BBS, articleDetailSummary_db, 200, nil
	}

	if result_b.Content == nil { // XXX possibly the article is deleted. Need to check error-code and mark the article as deleted.
		return nil, "", "", "", nil, 500, ErrNoArticle
	}

	// 7. parse article as content / commentsDBCS
	updateNanoTS = types.NowNanoTS()

	content, contentMD5, ip, host, bbs, signatureMD5, signatureDBCS, commentsDBCS := dbcs.ParseContent(result_b.Content, articleDetailSummary_db.ContentMD5)

	// update article
	// we need update-article-content be the 1st to upload,
	// because it's possible that there is no first-comments.
	// only article-content is guaranteed.

	err = UpdateArticleContentInfo(bboardID, articleID, content, contentMD5, ip, host, bbs, signatureMD5, signatureDBCS, updateNanoTS)

	if err != nil {
		return nil, "", "", "", nil, 500, err
	}

	if contentMD5 == "" {
		contentInfo, err := schema.GetArticleContentInfo(bboardID, articleID)
		if err != nil {
			return nil, "", "", "", nil, 500, err
		}
		content = contentInfo.Content
		ip = contentInfo.IP
		host = contentInfo.Host
		bbs = contentInfo.BBS
	}

	// 8. parse comments as firstComments and theRestComments
	firstComments, firstCommentsMD5, _, err := dbcs.ParseFirstComments(
		bboardID,
		articleID,
		ownerID,
		articleCreateTimeNanoTS,
		contentMTime,
		commentsDBCS,
		articleDetailSummary_db.FirstCommentsMD5,
	)

	// update first-comments
	// possibly err because the data is too old.
	// we don't need to queue and update content-mtime if the data is too old.
	err = tryUpdateFirstComments(firstComments, firstCommentsMD5, updateNanoTS, articleDetailSummary_db)
	if err != nil {
		//if failed update: we still send the content back.
		//(no updating the content in db,
		// so the data will be re-processed again next time).
		return content, ip, host, bbs, articleDetailSummary_db, 200, nil
	}

	// 9. enqueue and n_comments
	err = queue.QueueCommentDBCS(bboardID, articleID, ownerID, commentsDBCS, articleCreateTimeNanoTS, contentMTime, updateNanoTS)
	if err != nil {
		return content, ip, host, bbs, articleDetailSummary_db, 200, nil
	}

	if articleDetailSummary_db.NComments == 0 {
		articleDetailSummary_db.NComments = len(firstComments)
	}

	// everything is good, update content-mtime
	_ = schema.UpdateArticleContentMTime(bboardID, articleID, contentMTime)

	return content, ip, host, bbs, articleDetailSummary_db, 200, nil
}

func TryGetArticleContentInfoTooSoon(updateNanoTS types.NanoTS) bool {
	nowNanoTS := types.NowNanoTS()
	return nowNanoTS-updateNanoTS < GET_ARTICLE_CONTENT_INFO_TOO_SOON_NANO_TS
}

func TryGetArticleDetailSummary(userID bbs.UUserID, boardID bbs.BBoardID, articleID bbs.ArticleID, articleCreateTime pttbbstypes.Time4, c *gin.Context, isSystem bool) (articleDetailSummary *schema.ArticleDetailSummary, statusCode int, err error) {
	articleDetailSummary, err = schema.GetArticleDetailSummary(boardID, articleID)
	if err != nil { // something went wrong with db.
		return nil, 500, err
	}
	if articleDetailSummary != nil {
		return articleDetailSummary, 200, nil
	}

	// init startIdx
	articleSummary := &bbs.ArticleSummary{ArticleID: articleID, CreateTime: articleCreateTime}
	startIdx := bbs.SerializeArticleIdxStr(articleSummary)

	// backend load-general-articles
	theParams_b := &pttbbsapi.LoadGeneralArticlesParams{
		StartIdx:  startIdx,
		NArticles: 1,
		Desc:      false,
		IsSystem:  isSystem,
	}
	var result_b *pttbbsapi.LoadGeneralArticlesResult

	urlMap := map[string]string{
		"bid": string(boardID),
	}
	url := utils.MergeURL(urlMap, pttbbsapi.LOAD_GENERAL_ARTICLES_R)
	statusCode, err = utils.BackendGet(c, url, theParams_b, nil, &result_b)
	if err != nil || statusCode != 200 {
		return nil, statusCode, err
	}
	if len(result_b.Articles) == 0 {
		return nil, 500, ErrNoArticle
	}

	article_b := result_b.Articles[0]
	if article_b.ArticleID != articleID {
		return nil, 500, ErrNoArticle
	}

	// update to db
	updateNanoTS := types.NowNanoTS()
	articleSummaries_db, _, err := deserializeArticlesAndUpdateDB(userID, boardID, result_b.Articles, updateNanoTS)
	if err != nil {
		return nil, 500, err
	}

	articleSummary_db := articleSummaries_db[0]

	articleDetailSummary = &schema.ArticleDetailSummary{
		BBoardID:     boardID,
		ArticleID:    articleID,
		CreateTime:   articleSummary_db.CreateTime,
		MTime:        articleSummary_db.MTime,
		Recommend:    articleSummary_db.Recommend,
		Owner:        articleSummary_db.Owner,
		Title:        articleSummary_db.Title,
		Money:        articleSummary_db.Money,
		Class:        articleSummary_db.Class,
		Filemode:     articleSummary_db.Filemode,
		UpdateNanoTS: articleSummary_db.UpdateNanoTS,
	}

	return articleDetailSummary, 200, nil
}

func setUserReadArticle(content [][]*types.Rune, userID bbs.UUserID, articleID bbs.ArticleID, updateNanoTS types.NanoTS) {
	if content == nil {
		return
	}

	// user read article
	userReadArticle := &schema.UserReadArticle{
		UserID:       userID,
		ArticleID:    articleID,
		UpdateNanoTS: updateNanoTS,
	}
	_ = schema.UpdateUserReadArticle(userReadArticle)
}
