package api

import (
	"sync"
	"testing"

	"github.com/Ptt-official-app/go-openbbsmiddleware/apitypes"
	"github.com/Ptt-official-app/go-openbbsmiddleware/schema"
	"github.com/Ptt-official-app/go-openbbsmiddleware/types"
	"github.com/Ptt-official-app/go-pttbbs/bbs"
	"github.com/Ptt-official-app/go-pttbbs/testutil"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TestLoadUserArticles(t *testing.T) {
	setupTest()
	defer teardownTest()

	boardSummaries_b := []*bbs.BoardSummary{testBoardSummaryWhoAmI_b, testBoardSummarySYSOP_b}
	_, _, _ = deserializeBoardsAndUpdateDB("SYSOP", boardSummaries_b, 123456890000000000)

	update0 := &schema.UserReadArticle{UserID: "SYSOP", BoardID: "10_WhoAmI", ArticleID: "19bWBI4Z", UpdateNanoTS: types.Time8(1534567891).ToNanoTS()}
	update1 := &schema.UserReadArticle{UserID: "SYSOP", BoardID: "10_WhoAmI", ArticleID: "1VrooM21", UpdateNanoTS: types.Time8(1234567800).ToNanoTS()}

	_, _ = schema.UserReadArticle_c.Update(update0, update0)
	_, _ = schema.UserReadArticle_c.Update(update1, update1)

	paramsLoadGeneralArticles := NewLoadGeneralArticlesParams()
	pathLoadGeneralArticles := &LoadGeneralArticlesPath{FBoardID: "WhoAmI"}
	LoadGeneralArticles("localhost", "SYSOP", paramsLoadGeneralArticles, pathLoadGeneralArticles, &gin.Context{})

	paramsLoadGeneralArticles = NewLoadGeneralArticlesParams()
	pathLoadGeneralArticles = &LoadGeneralArticlesPath{FBoardID: "SYSOP"}
	LoadGeneralArticles("localhost", "SYSOP", paramsLoadGeneralArticles, pathLoadGeneralArticles, &gin.Context{})

	articleSummary, _ := schema.GetArticleSummary("10_WhoAmI", "19bWBI4Z")
	logrus.Infof("TestLoadUserATestLoadUserArticles: articleSummary: %v", articleSummary)

	articleSummary, _ = schema.GetArticleSummary("10_WhoAmI", "1VrooM21")
	logrus.Infof("TestLoadUserATestLoadUserArticles: articleSummary: %v", articleSummary)

	articleSummary, _ = schema.GetArticleSummary("1_SYSOP", "1VrooM21")
	logrus.Infof("TestLoadUserATestLoadUserArticles: articleSummary: %v", articleSummary)

	params0 := NewUserArticlesParams()
	path0 := &LoadUserArticlesPath{
		UserID: "teemo",
	}

	expectedResult0 := &LoadUserArticlesResult{
		List: []*apitypes.ArticleSummary{
			{
				FBoardID:   apitypes.FBoardID("SYSOP"),
				ArticleID:  apitypes.FArticleID("M.1607937176.A.081"),
				IsDeleted:  false,
				CreateTime: types.Time8(1607937176),
				MTime:      types.Time8(1607937100),
				Recommend:  10,
				Owner:      "teemo",
				Title:      "再來呢？～",
				Class:      "問題",
				Money:      12,
				Filemode:   0,
				URL:        "http://localhost:3457/bbs/SYSOP/M.1607937176.A.081.html",
				Read:       false,
				Idx:        "1607937176@1VrooO21",
			},
			{
				FBoardID:   apitypes.FBoardID("WhoAmI"),
				ArticleID:  apitypes.FArticleID("M.1607937174.A.081"),
				IsDeleted:  false,
				CreateTime: types.Time8(1607937174),
				MTime:      types.Time8(1607937100),
				Recommend:  3,
				Owner:      "teemo",
				Title:      "再來呢？～",
				Class:      "問題",
				Money:      12,
				Filemode:   0,
				URL:        "http://localhost:3457/bbs/WhoAmI/M.1607937174.A.081.html",
				Read:       false,
				Idx:        "1607937174@1VrooM21",
			},
			{
				FBoardID:   apitypes.FBoardID("SYSOP"),
				ArticleID:  apitypes.FArticleID("M.1234567892.A.123"),
				IsDeleted:  false,
				CreateTime: types.Time8(1234567892),
				MTime:      types.Time8(1234567889),
				Recommend:  24,
				Owner:      "teemo",
				Title:      "然後呢？～",
				Class:      "問題",
				Money:      3,
				Filemode:   0,
				URL:        "http://localhost:3457/bbs/SYSOP/M.1234567892.A.123.html",
				Read:       false,
				Idx:        "1234567892@19bWBK4Z",
			},
		},
	}

	params1 := NewUserArticlesParams()
	path1 := &LoadUserArticlesPath{
		UserID: "okcool",
	}

	expectedResult1 := &LoadUserArticlesResult{
		List: []*apitypes.ArticleSummary{
			{
				FBoardID:   apitypes.FBoardID("WhoAmI"),
				ArticleID:  apitypes.FArticleID("M.1234567890.A.123"),
				IsDeleted:  false,
				CreateTime: types.Time8(1234567890),
				MTime:      types.Time8(1234567889),
				Recommend:  8,
				Owner:      "okcool",
				Title:      "然後呢？～",
				Class:      "問題",
				Money:      3,
				Filemode:   0,
				URL:        "http://localhost:3457/bbs/WhoAmI/M.1234567890.A.123.html",
				Read:       true,
				Idx:        "1234567890@19bWBI4Z",
			},
		},
	}

	params2 := NewUserArticlesParams()
	path2 := &LoadUserArticlesPath{
		UserID: "nonexists",
	}
	expectedResult2 := &LoadUserArticlesResult{
		List: []*apitypes.ArticleSummary{},
	}

	params3 := &LoadUserArticlesParams{
		Descending: true,
		Max:        2,
	}
	path3 := &LoadUserArticlesPath{
		UserID: "teemo",
	}

	expectedResult3 := &LoadUserArticlesResult{
		List: []*apitypes.ArticleSummary{
			{
				FBoardID:   apitypes.FBoardID("SYSOP"),
				ArticleID:  apitypes.FArticleID("M.1607937176.A.081"),
				IsDeleted:  false,
				CreateTime: types.Time8(1607937176),
				MTime:      types.Time8(1607937100),
				Recommend:  10,
				Owner:      "teemo",
				Title:      "再來呢？～",
				Class:      "問題",
				Money:      12,
				Filemode:   0,
				URL:        "http://localhost:3457/bbs/SYSOP/M.1607937176.A.081.html",
				Read:       false,
				Idx:        "1607937176@1VrooO21",
			},
			{
				FBoardID:   apitypes.FBoardID("WhoAmI"),
				ArticleID:  apitypes.FArticleID("M.1607937174.A.081"),
				IsDeleted:  false,
				CreateTime: types.Time8(1607937174),
				MTime:      types.Time8(1607937100),
				Recommend:  3,
				Owner:      "teemo",
				Title:      "再來呢？～",
				Class:      "問題",
				Money:      12,
				Filemode:   0,
				URL:        "http://localhost:3457/bbs/WhoAmI/M.1607937174.A.081.html",
				Read:       false,
				Idx:        "1607937174@1VrooM21",
			},
		},
		NextIdx: "1234567892@19bWBK4Z",
	}

	params4 := &LoadUserArticlesParams{
		StartIdx:   "1607937174@1VrooM21",
		Descending: true,
		Max:        1,
	}
	path4 := &LoadUserArticlesPath{
		UserID: "teemo",
	}

	expectedResult4 := &LoadUserArticlesResult{
		List: []*apitypes.ArticleSummary{
			{
				FBoardID:   apitypes.FBoardID("WhoAmI"),
				ArticleID:  apitypes.FArticleID("M.1607937174.A.081"),
				IsDeleted:  false,
				CreateTime: types.Time8(1607937174),
				MTime:      types.Time8(1607937100),
				Recommend:  3,
				Owner:      "teemo",
				Title:      "再來呢？～",
				Class:      "問題",
				Money:      12,
				Filemode:   0,
				URL:        "http://localhost:3457/bbs/WhoAmI/M.1607937174.A.081.html",
				Read:       false,
				Idx:        "1607937174@1VrooM21",
			},
		},
		NextIdx: "1234567892@19bWBK4Z",
	}

	type args struct {
		remoteAddr string
		userID     bbs.UUserID
		params     interface{}
		path       interface{}
		c          *gin.Context
	}
	tests := []struct {
		name               string
		args               args
		expectedResult     interface{}
		expectedStatusCode int
		wantErr            bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				remoteAddr: "127.0.0.1",
				userID:     "SYSOP",
				params:     params0,
				path:       path0,
			},
			expectedResult:     expectedResult0,
			expectedStatusCode: 200,
		},
		{
			args: args{
				remoteAddr: "127.0.0.1",
				userID:     "SYSOP",
				params:     params1,
				path:       path1,
			},
			expectedResult:     expectedResult1,
			expectedStatusCode: 200,
		},
		{
			args: args{
				remoteAddr: "127.0.0.1",
				userID:     "SYSOP",
				params:     params2,
				path:       path2,
			},
			expectedResult:     expectedResult2,
			expectedStatusCode: 200,
		},
		{
			name: "limit to 2",
			args: args{
				remoteAddr: "127.0.0.1",
				userID:     "SYSOP",
				params:     params3,
				path:       path3,
			},
			expectedResult:     expectedResult3,
			expectedStatusCode: 200,
		},
		{
			name: "with start-idx",
			args: args{
				remoteAddr: "127.0.0.1",
				userID:     "SYSOP",
				params:     params4,
				path:       path4,
			},
			expectedResult:     expectedResult4,
			expectedStatusCode: 200,
		},
	}
	var wg sync.WaitGroup
	for _, tt := range tests {
		wg.Add(1)
		t.Run(tt.name, func(t *testing.T) {
			defer wg.Done()
			gotResult, gotStatusCode, err := LoadUserArticles(tt.args.remoteAddr, tt.args.userID, tt.args.params, tt.args.path, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadUserArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			testutil.TDeepEqual(t, "got", gotResult, tt.expectedResult)
			if gotStatusCode != tt.expectedStatusCode {
				t.Errorf("LoadUserArticles() gotStatusCode = %v, want %v", gotStatusCode, tt.expectedStatusCode)
			}
		})
		wg.Wait()
	}
}
