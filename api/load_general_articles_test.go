package api

import (
	"net/http"
	"net/url"
	"sync"
	"testing"

	"github.com/Ptt-official-app/go-openbbsmiddleware/apitypes"
	"github.com/Ptt-official-app/go-openbbsmiddleware/schema"
	"github.com/Ptt-official-app/go-openbbsmiddleware/types"
	"github.com/Ptt-official-app/go-pttbbs/bbs"
	"github.com/Ptt-official-app/go-pttbbs/testutil"
	"github.com/gin-gonic/gin"
)

func TestLoadGeneralArticles(t *testing.T) {
	setupTest()
	defer teardownTest()

	defer func() {
		schema.UserReadArticle_c.Drop()
		schema.UserReadBoard_c.Drop()
	}()

	update0 := &schema.UserReadArticle{UserID: "SYSOP", ArticleID: "19bWBI4Zokcool", UpdateNanoTS: types.Time8(1534567891).ToNanoTS()}
	update1 := &schema.UserReadArticle{UserID: "SYSOP", ArticleID: "1VrooM21teemo", UpdateNanoTS: types.Time8(1234567800).ToNanoTS()}

	_, _ = schema.UserReadArticle_c.Update(update0, update0)
	_, _ = schema.UserReadArticle_c.Update(update1, update1)

	params := NewLoadGeneralArticlesParams()
	path := &LoadGeneralArticlesPath{BBoardID: "1_test1"}
	expectedResult := &LoadGeneralArticlesResult{
		List: []*apitypes.ArticleSummary{
			{
				BBoardID:   bbs.BBoardID("1_test1"),
				ArticleID:  bbs.ArticleID("1VrooM21teemo"),
				IsDeleted:  false,
				CreateTime: types.Time8(1607937174),
				MTime:      types.Time8(1607937100),
				Recommend:  3,
				Owner:      "teemo",
				Title:      "再來呢？～",
				Class:      "問題",
				Money:      12,
				Filemode:   0,
				URL:        "http://localhost:3457/bbs/1_test1/M.1607937174.A.081.html",
				Read:       false,
			},
			{
				BBoardID:   bbs.BBoardID("1_test1"),
				ArticleID:  bbs.ArticleID("19bWBI4Zokcool"),
				IsDeleted:  false,
				CreateTime: types.Time8(1234567890),
				MTime:      types.Time8(1234567889),
				Recommend:  8,
				Owner:      "okcool",
				Title:      "然後呢？～",
				Class:      "問題",
				Money:      3,
				Filemode:   0,
				URL:        "http://localhost:3457/bbs/1_test1/M.1234567890.A.123.html",
				Read:       true,
			},
		},
		NextIdx:        "1234560000@19bUG021okcool",
		NextCreateTime: 1234560000,
	}

	c := &gin.Context{}
	c.Request = &http.Request{URL: &url.URL{Path: "/api/board/1_test/articles"}}
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
			args:               args{remoteAddr: "localhost", userID: "SYSOP", params: params, path: path, c: &gin.Context{}},
			expectedResult:     expectedResult,
			expectedStatusCode: 200,
		},
	}

	var wg sync.WaitGroup
	for _, tt := range tests {
		wg.Add(1)
		t.Run(tt.name, func(t *testing.T) {
			defer wg.Done()
			gotResult, gotStatusCode, err := LoadGeneralArticles(tt.args.remoteAddr, tt.args.userID, tt.args.params, tt.args.path, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadGeneralArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			testutil.TDeepEqual(t, "result", gotResult, tt.expectedResult)

			if gotStatusCode != tt.expectedStatusCode {
				t.Errorf("LoadGeneralArticles() gotStatusCode = %v, want %v", gotStatusCode, tt.expectedStatusCode)
			}

		})
	}
	wg.Wait()
}
