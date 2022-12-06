package posting

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gochan-org/gochan/pkg/config"
	"github.com/gochan-org/gochan/pkg/gcsql"
	"github.com/gochan-org/gochan/pkg/gctemplates"
	"github.com/gochan-org/gochan/pkg/gcutil"
	"github.com/gochan-org/gochan/pkg/serverutil"
	"github.com/rs/zerolog"
)

func showBanpage(ban *gcsql.IPBan, post *gcsql.Post, postBoard *gcsql.Board, writer http.ResponseWriter, request *http.Request) {
	banPageBuffer := bytes.NewBufferString("")
	err := serverutil.MinifyTemplate(gctemplates.Banpage, map[string]interface{}{
		"systemCritical": config.GetSystemCriticalConfig(),
		"siteConfig":     config.GetSiteConfig(),
		"boardConfig":    config.GetBoardConfig(postBoard.Dir),
		"ban":            ban,
		"board":          postBoard,
		"permanent":      ban.Permanent,
		"expires":        ban.ExpiresAt,
	}, banPageBuffer, "text/html")
	if err != nil {
		gcutil.LogError(err).
			Str("IP", post.IP).
			Str("building", "minifier").
			Str("template", "banpage.html").Send()
		serverutil.ServeErrorPage(writer, "Error minifying page: "+err.Error())
		return
	}
	writer.Write(banPageBuffer.Bytes())
	gcutil.LogWarning().
		Str("IP", post.IP).
		Str("boardDir", postBoard.Dir).
		Msg("Rejected post from banned IP")
}

// checks the post for spam. It returns true if a ban page or an error page was served (causing MakePost() to return)
func checkIpBan(post *gcsql.Post, postBoard *gcsql.Board, writer http.ResponseWriter, request *http.Request) bool {
	ipBan, err := gcsql.CheckIPBan(post.IP, postBoard.ID)
	if err != nil {
		gcutil.LogError(err).
			Str("IP", post.IP).
			Str("boardDir", postBoard.Dir).
			Msg("Error getting IP banned status")
		serverutil.ServeErrorPage(writer, "Error getting ban info"+err.Error())
		return true
	}
	if ipBan == nil {
		return false // ip is not banned and there were no errors, keep going
	}
	// IP is banned
	showBanpage(ipBan, post, postBoard, writer, request)
	return true
}

func checkUsernameBan(formName string, post *gcsql.Post, postBoard *gcsql.Board, writer http.ResponseWriter, request *http.Request) bool {
	if formName == "" {
		return false
	}

	nameBan, err := gcsql.CheckNameBan(formName, postBoard.ID)
	if err != nil {
		gcutil.LogError(err).
			Str("IP", post.IP).
			Str("name", formName).
			Str("boardDir", postBoard.Dir).
			Msg("Error getting name banned status")
		serverutil.ServeErrorPage(writer, "Error getting name ban info")
		return true
	}
	if nameBan == nil {
		return false // name is not banned
	}
	serverutil.ServeError(writer, "Name or tripcode not allowed", serverutil.IsRequestingJSON(request), map[string]interface{}{})
	gcutil.LogWarning().
		Str("IP", post.IP).
		Str("boardDir", postBoard.Dir).
		Str("name", post.Name).
		Str("tripcode", post.Tripcode).
		Msg("Rejected post with banned name/tripcode")
	return true
}

func checkFilenameBan(upload *gcsql.Upload, post *gcsql.Post, postBoard *gcsql.Board, writer http.ResponseWriter, request *http.Request) bool {
	filenameBan, err := gcsql.CheckFilenameBan(upload.OriginalFilename, postBoard.ID)
	if err != nil {
		gcutil.LogError(err).
			Str("IP", post.IP).
			Str("filename", upload.OriginalFilename).
			Str("boardDir", postBoard.Dir).
			Msg("Error getting name banned status")
		serverutil.ServeErrorPage(writer, "Error getting filename ban info")
		return true
	}
	if filenameBan == nil {
		return false
	}
	serverutil.ServeError(writer, "Filename not allowed", serverutil.IsRequestingJSON(request), map[string]interface{}{})
	gcutil.LogWarning().
		Str("originalFilename", upload.OriginalFilename).
		Msg("File rejected for having a banned filename")
	return true
}

func checkChecksumBan(upload *gcsql.Upload, post *gcsql.Post, postBoard *gcsql.Board, writer http.ResponseWriter, request *http.Request) bool {
	fileBan, err := gcsql.CheckFileChecksumBan(upload.Checksum, postBoard.ID)
	if err != nil {
		gcutil.LogError(err).
			Str("IP", post.IP).
			Str("boardDir", postBoard.Dir).
			Str("checksum", upload.Checksum).
			Msg("Error getting file checksum ban status")
		serverutil.ServeErrorPage(writer, "Error processing file: "+err.Error())
		return true
	}
	if fileBan == nil {
		return false
	}
	serverutil.ServeError(writer, "File not allowed", serverutil.IsRequestingJSON(request), map[string]interface{}{})
	gcutil.LogWarning().
		Str("originalFilename", upload.OriginalFilename).
		Str("checksum", upload.Checksum).
		Msg("File rejected for having a banned checksum")
	return true
}

func handleAppeal(writer http.ResponseWriter, request *http.Request, errEv *zerolog.Event) {
	banIDstr := request.FormValue("banid")
	if banIDstr == "" {
		errEv.Caller().Msg("Appeal sent without banid field")
		serverutil.ServeErrorPage(writer, "Missing banid value")
		return
	}
	appealMsg := request.FormValue("appealmsg")
	if appealMsg == "" {
		errEv.Caller().Msg("Missing appealmsg value")
		serverutil.ServeErrorPage(writer, "Missing or empty appeal")
		return
	}
	banID, err := strconv.Atoi(banIDstr)
	if err != nil {
		errEv.Err(err).
			Str("banIDstr", banIDstr).Caller().Send()
		serverutil.ServeErrorPage(writer, fmt.Sprintf("Invalid banid value %q", banIDstr))
		return
	}
	errEv.Int("banID", banID)
	ban, err := gcsql.GetIPBanByID(banID)
	if err != nil {
		errEv.Err(err).
			Caller().Send()
		serverutil.ServeErrorPage(writer, "Error getting ban info: "+err.Error())
		return
	}
	if ban == nil {
		errEv.Caller().Msg("GetIPBanByID returned a nil ban (presumably not banned)")
		serverutil.ServeErrorPage(writer, fmt.Sprintf("Invalid banid %d", banID))
		return
	}
	if ban.IP != gcutil.GetRealIP(request) {
		errEv.Caller().
			Str("banIP", ban.IP).
			Msg("User tried to appeal a ban from a different IP")
		serverutil.ServeErrorPage(writer, fmt.Sprintf("Invalid banid %d", banID))
		return
	}
	if !ban.IsActive {
		errEv.Caller().Msg("Requested ban is not active")
		serverutil.ServeErrorPage(writer, "Requested ban is not active")
		return
	}
	if !ban.CanAppeal {
		errEv.Caller().Msg("Rejected appeal submission, appeals denied for this ban")
		serverutil.ServeErrorPage(writer, "You can not appeal this ban")
	}
	if ban.AppealAt.After(time.Now()) {
		errEv.Caller().
			Time("appealAt", ban.AppealAt).
			Msg("Rejected appeal submission, can't appeal yet")
		serverutil.ServeErrorPage(writer, "You are not able to appeal this ban until "+ban.AppealAt.Format(config.GetBoardConfig("").DateTimeFormat))
	}
	if err = ban.Appeal(appealMsg); err != nil {
		errEv.Err(err).
			Str("appealMsg", appealMsg).
			Caller().Msg("Unable to submit appeal")
		serverutil.ServeErrorPage(writer, "Unable to submit appeal")
		return
	}
	board := request.FormValue("board")
	gcutil.LogInfo().
		Str("IP", gcutil.GetRealIP(request)).
		Int("banID", banID).
		Str("board", board).
		Msg("Appeal submitted")
	http.Redirect(writer, request, config.WebPath(request.FormValue("board")), http.StatusFound)
}
