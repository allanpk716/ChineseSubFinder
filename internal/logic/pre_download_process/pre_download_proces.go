package pre_download_process

import (
	"errors"
	"fmt"
	"time"

	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/a4k"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/global_value"

	"github.com/allanpk716/ChineseSubFinder/internal/logic/file_downloader"
	subSupplier "github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/assrt"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/csf"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/shooter"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/subhd"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_supplier/xunlei"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_folder"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/notify_center"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/settings"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/something_static"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/url_connectedness_helper"
	common2 "github.com/allanpk716/ChineseSubFinder/internal/types/common"
	"github.com/sirupsen/logrus"
)

type PreDownloadProcess struct {
	stageName string
	gError    error

	settings       *settings.Settings
	log            *logrus.Logger
	fileDownloader *file_downloader.FileDownloader
	SubSupplierHub *subSupplier.SubSupplierHub
}

func NewPreDownloadProcess(_fileDownloader *file_downloader.FileDownloader) *PreDownloadProcess {

	preDownloadProcess := PreDownloadProcess{
		fileDownloader: _fileDownloader,
		log:            _fileDownloader.Log,
		settings:       _fileDownloader.Settings,
	}
	return &preDownloadProcess
}

func (p *PreDownloadProcess) Init() *PreDownloadProcess {

	if p.gError != nil {
		p.log.Infoln("Skip PreDownloadProcess.Init()")
		return p
	}
	p.stageName = stageNameInit
	defer func() {
		p.log.Infoln("PreDownloadProcess.Init() End")
	}()
	p.log.Infoln("PreDownloadProcess.Init() Start...")

	// ------------------------------------------------------------------------
	// ???????????????????????????
	notify_center.Notify = notify_center.NewNotifyCenter(p.log, p.settings.DeveloperSettings.BarkServerAddress)
	// ??????????????????
	notify_center.Notify.Clear()
	// ------------------------------------------------------------------------
	// ???????????????
	if global_value.LiteMode() == false {

		nowTT := time.Now()
		nowTimeFileNamePrix := fmt.Sprintf("%d%d%d", nowTT.Year(), nowTT.Month(), nowTT.Day())
		updateTimeString, code, err := something_static.GetCodeFromWeb(p.log, nowTimeFileNamePrix, p.fileDownloader)
		if err != nil {
			notify_center.Notify.Add("GetSubhdCode", "GetCodeFromWeb,"+err.Error())
			p.log.Errorln("something_static.GetCodeFromWeb", err)
			p.log.Errorln("Skip Subhd download")
			// ?????????????????????
			common2.SubhdCode = ""
		} else {

			// ???????????????????????????????????????????????????????????????????????????
			codeTime, err := time.Parse("2006-01-02", updateTimeString)
			if err != nil {
				p.log.Errorln("something_static.GetCodeFromWeb.time.Parse", err)
				// ?????????????????????
				common2.SubhdCode = ""
			} else {

				nowTime := time.Now()
				if codeTime.YearDay() != nowTime.YearDay() {
					// ?????????????????????
					common2.SubhdCode = ""
					p.log.Warningln("something_static.GetCodeFromWeb, GetCodeTime:", updateTimeString, "NowTime:", time.Now().String(), "Skip")
				} else {
					p.log.Infoln("GetCode", updateTimeString, code)
					common2.SubhdCode = code
				}
			}
		}
	}
	// ------------------------------------------------------------------------
	// ??????????????????????????????????????????
	if p.settings.SpeedDevMode == true {

		p.SubSupplierHub = subSupplier.NewSubSupplierHub(
			csf.NewSupplier(p.fileDownloader),
		)
	} else {

		p.SubSupplierHub = subSupplier.NewSubSupplierHub(
			//zimuku.NewSupplier(p.fileDownloader),
			xunlei.NewSupplier(p.fileDownloader),
			shooter.NewSupplier(p.fileDownloader),
			a4k.NewSupplier(p.fileDownloader),
		)

		if p.settings.ExperimentalFunction.ShareSubSettings.ShareSubEnabled == true {
			// ?????????????????????????????????????????????????????????????????????
			p.SubSupplierHub.AddSubSupplier(csf.NewSupplier(p.fileDownloader))
		}

		if p.settings.SubtitleSources.AssrtSettings.Enabled == true &&
			p.settings.SubtitleSources.AssrtSettings.Token != "" {
			// ??????????????? ASSRt ???????????????????????????
			p.SubSupplierHub.AddSubSupplier(assrt.NewSupplier(p.fileDownloader))
		}

		if global_value.LiteMode() == false {
			// ???????????? Lite ??????????????????????????????????????????
			if common2.SubhdCode != "" {
				// ???????????? code ??????????????????????????????????????????
				p.SubSupplierHub.AddSubSupplier(subhd.NewSupplier(p.fileDownloader))
			}
		}
	}
	// ------------------------------------------------------------------------
	// ?????????????????? rod ????????????
	err := my_folder.ClearRodTmpRootFolder()
	if err != nil {
		p.gError = errors.New("ClearRodTmpRootFolder " + err.Error())
		return p
	}

	p.log.Infoln("ClearRodTmpRootFolder Done")

	return p
}

func (p *PreDownloadProcess) Check() *PreDownloadProcess {

	if p.gError != nil {
		p.log.Infoln("Skip PreDownloadProcess.Check()")
		return p
	}
	p.stageName = stageNameCheck
	defer func() {
		p.log.Infoln("PreDownloadProcess.Check() End")
	}()
	p.log.Infoln("PreDownloadProcess.Check() Start...")
	// ------------------------------------------------------------------------
	// ??????????????????
	if p.settings.AdvancedSettings.ProxySettings.UseProxy == false {

		p.log.Infoln("UseHttpProxy = false")
		// ???????????????????????????????????????????????? baidu ?????????????????????????????????
		proxyStatus, proxySpeed, err := url_connectedness_helper.UrlConnectednessTest(url_connectedness_helper.BaiduUrl, "")
		if err != nil {
			p.log.Errorln(errors.New("UrlConnectednessTest Target Site " + url_connectedness_helper.BaiduUrl + ", " + err.Error()))
		} else {
			p.log.Infoln("UrlConnectednessTest Target Site", url_connectedness_helper.BaiduUrl, "Speed:", proxySpeed, "ms,", "Status:", proxyStatus)
		}
	} else {

		p.log.Infoln("UseHttpProxy By:", p.settings.AdvancedSettings.ProxySettings.UseWhichProxyProtocol)
		// ???????????????????????????????????????????????? google ?????????????????????????????????
		proxyStatus, proxySpeed, err := url_connectedness_helper.UrlConnectednessTest(url_connectedness_helper.GoogleUrl, p.settings.AdvancedSettings.ProxySettings.GetLocalHttpProxyUrl())
		if err != nil {
			p.log.Errorln(errors.New("UrlConnectednessTest Target Site " + url_connectedness_helper.GoogleUrl + ", " + err.Error()))
		} else {
			p.log.Infoln("UrlConnectednessTest Target Site", url_connectedness_helper.GoogleUrl, "Speed:", proxySpeed, "ms,", "Status:", proxyStatus)
		}
	}
	// ------------------------------------------------------------------------
	// ??????????????????????????????????????????????????????????????????
	p.SubSupplierHub.CheckSubSiteStatus()
	// ------------------------------------------------------------------------
	// ???????????????????????????
	if len(p.settings.CommonSettings.MoviePaths) < 1 {
		p.log.Warningln("MoviePaths not set, len == 0")
	}
	if len(p.settings.CommonSettings.SeriesPaths) < 1 {
		p.log.Warningln("SeriesPaths not set, len == 0")
	}
	for i, path := range p.settings.CommonSettings.MoviePaths {
		if my_util.IsDir(path) == false {
			p.log.Errorln("MovieFolder not found Index", i, "--", path)
		} else {
			p.log.Infoln("MovieFolder Index", i, "--", path)
		}
	}
	for i, path := range p.settings.CommonSettings.SeriesPaths {
		if my_util.IsDir(path) == false {
			p.log.Errorln("SeriesPaths not found Index", i, "--", path)
		} else {
			p.log.Infoln("SeriesPaths Index", i, "--", path)
		}
	}
	// ------------------------------------------------------------------------
	// ??????????????? Emby ????????????????????????

	return p
}

func (p *PreDownloadProcess) Wait() error {
	defer func() {
		p.log.Infoln("PreDownloadProcess.Wait() Done.")
	}()
	if p.gError != nil {
		outErrString := "PreDownloadProcess.Wait() Get Error, " + "stageName:" + p.stageName + " -- " + p.gError.Error()
		p.log.Errorln(outErrString)
		return errors.New(outErrString)
	} else {
		return nil
	}
}

const (
	stageNameInit  = "Init"
	stageNameCheck = "Check"
)
