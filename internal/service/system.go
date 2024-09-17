package service

import (
	"context"
	"strings"

	"github.com/0x0FACED/link-saver-api/internal/domain/models"
	"github.com/0x0FACED/proto-files/link_service/gen"
	"github.com/gocolly/colly"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *LinkService) SaveLink(ctx context.Context, req *gen.SaveLinkRequest) (*gen.SaveLinkResponse, error) {
	s.logger.Info("Received link",
		zap.Int64("user", req.UserId),
		zap.String("desc", req.Description),
		zap.String("url", req.OriginalUrl),
	)
	var link *models.Link

	page := s.rod.MustPage(req.OriginalUrl)

	s.colly.OnHTML("script[src]", func(e *colly.HTMLElement) {
		jsPath := e.Attr("src")
		if strings.HasPrefix(jsPath, "http://") || strings.HasPrefix(jsPath, "https://") {
			s.logger.Debug("External script link, skipping", zap.String("url", jsPath))
			return
		}

		jsURL := e.Request.AbsoluteURL(jsPath)
		jsContent := s.fetchResourceContent(jsURL, e)
		if jsContent != nil {
			res := &models.Resource{
				Name:    hashResName(getRelativePath(jsURL), models.ScriptType),
				Content: jsContent,
				Type:    models.ScriptType,
			}
			err := s.saveResource(res)
			if err != nil {
				s.logger.Debug("Failed to save script", zap.Any("script", res.Name))
			}
			newPath := getResourceURL(s.cfg.BaseURL, "script", res.Name)
			e.DOM.SetAttr("src", newPath)
		}
	})

	s.colly.OnHTML("link[rel='stylesheet']", func(e *colly.HTMLElement) {
		cssPath := e.Attr("href")
		if strings.HasPrefix(cssPath, "http://") || strings.HasPrefix(cssPath, "https://") {
			s.logger.Debug("External stylesheet link, skipping", zap.String("url", cssPath))
			return
		}

		cssURL := e.Request.AbsoluteURL(cssPath)
		cssContent := s.fetchResourceContent(cssURL, e)
		if cssContent != nil {
			res := &models.Resource{
				Name:    hashResName(getRelativePath(cssURL), models.CSSType),
				Content: cssContent,
				Type:    models.CSSType,
			}
			err := s.saveResource(res)
			if err != nil {
				s.logger.Debug("Failed to save CSS", zap.Any("css", res.Name))
			}
			newPath := getResourceURL(s.cfg.BaseURL, "css", res.Name)
			e.DOM.SetAttr("href", newPath)
		}
	})

	s.colly.OnHTML("img[src]", func(e *colly.HTMLElement) {
		imgPath := e.Attr("src")
		if strings.HasPrefix(imgPath, "http://") || strings.HasPrefix(imgPath, "https://") {
			s.logger.Debug("External image link, skipping", zap.String("url", imgPath))
			return
		}

		imgURL := e.Request.AbsoluteURL(imgPath)
		imgContent := s.fetchResourceContent(imgURL, e)
		if imgContent != nil {
			res := &models.Resource{
				Name:    hashResName(getRelativePath(imgURL), models.ImageType),
				Content: imgContent,
				Type:    models.ImageType,
			}
			err := s.saveResource(res)
			if err != nil {
				s.logger.Debug("Failed to save image", zap.Any("image", res.Name))
			}
			newPath := getResourceURL(s.cfg.BaseURL, "image", res.Name)
			e.DOM.SetAttr("src", newPath)
		}
	})

	page.MustWaitLoad()

	// Test for wiki kangaroos
	page.MustEval(`
		fetch('/w/load.php?lang=en&modules=ext.cite.styles%7Cext.uls.interlanguage%7Cext.visualEditor.desktopArticleTarget.noscript%7Cext.wikimediaBadges%7Cext.wikimediamessages.styles%7Cjquery.makeCollapsible.styles%7Cskins.vector.icons%2Cstyles%7Cskins.vector.search.codex.styles%7Cwikibase.client.init&only=styles&skin=vector-2022')
			.then(response => response.text())
			.then(text => {
				document.querySelector('head').insertAdjacentHTML('beforeend', text);
			});
	`)

	updatedHTML, err := page.HTML()
	if err != nil {
		s.logger.Error("Error retrieving updated HTML", zap.Error(err))
		return &gen.SaveLinkResponse{Success: false, Message: "Error processing page"}, err
	}

	link = &models.Link{
		OriginalURL: req.OriginalUrl,
		UserID:      req.UserId,
		Description: req.Description,
		Content:     []byte(updatedHTML),
	}
	s.logger.Info("Visited link", zap.String("url", link.OriginalURL))

	err = s.saveLink(ctx, link)
	if err != nil {
		s.logger.Error("Failed to save to db", zap.Error(err))
		return &gen.SaveLinkResponse{Success: false, Message: "Not saved, already exists"}, status.Error(codes.AlreadyExists, "link already exists")
	}
	s.logger.Debug("Link successfully saved to db")

	return &gen.SaveLinkResponse{Success: true, Message: "Successfully saved"}, nil
}

func (s *LinkService) DeleteLink(ctx context.Context, req *gen.DeleteLinkRequest) (*gen.DeleteLinkResponse, error) {
	s.logger.Info("New req GetAllLinks()",
		zap.Int64("link_id", int64(req.LinkId)),
	)
	original, id, err := s.db.DeleteLink(ctx, int(req.LinkId))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Link not found: %v", err)
	}

	err = s.redis.DeleteLink(ctx, id, original)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete link from Redis: %v", err)
	}

	return &gen.DeleteLinkResponse{Success: true, Message: "Successfully deleted"}, nil
}

func (s *LinkService) GetLinks(ctx context.Context, req *gen.GetLinksRequest) (*gen.GetLinksResponse, error) {
	s.logger.Info("New req GetLinks()",
		zap.Int64("user", req.UserId),
		zap.String("desc", req.Description),
	)

	links, err := s.db.GetLinksByTelegramIDDesc(ctx, req.UserId, req.Description)
	if err != nil {
		s.logger.Error("Failed to get links by username and desc",
			zap.Int64("user", req.UserId),
			zap.String("desc", req.Description),
		)
		return &gen.GetLinksResponse{Links: nil}, status.Errorf(codes.Internal, "Failed to get links: %v", err)
	}
	s.logger.Debug("Found links", zap.Int64("user", req.UserId), zap.Any("links", links))

	return &gen.GetLinksResponse{Links: links}, nil
}

func (s *LinkService) GetLink(ctx context.Context, req *gen.GetLinkRequest) (*gen.GetLinkResponse, error) {
	s.logger.Info("New req GetLink()",
		zap.Int64("user", req.UserId),
		zap.String("desc", req.Description),
		zap.Int32("url_id", req.UrlId),
	)

	l, err := s.db.GetLinkByID(ctx, int(req.UrlId))
	if err != nil {
		s.logger.Error("Failed to get link by id from Postgres",
			zap.Int64("user", req.UserId),
			zap.String("desc", req.Description),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.NotFound, "Link not found: %v", err)
	}

	redisLink, err := s.redis.GetLink(ctx, req.UserId, l.OriginalURL)
	if err != nil && err != redis.Nil {
		s.logger.Error("Failed to get link from Redis",
			zap.Int64("user", req.UserId),
			zap.String("desc", req.Description),
			zap.Error(err),
		)
		//return nil, status.Errorf(codes.Internal, "Failed to get link from Redis: %v", err)
	}

	if redisLink != nil {
		s.logger.Debug("Link found in Redis",
			zap.Int64("user", req.UserId),
			zap.String("desc", req.Description),
		)

		fullURL := getFullLink(s.cfg.BaseURL, l.UserID, redisLink.Link)
		s.logger.Debug("Generated Full Link",
			zap.String("full_link", fullURL),
		)
		return &gen.GetLinkResponse{GeneratedUrl: fullURL}, nil
	}

	generatedLink := hash(l.UserID, l.OriginalURL)

	err = s.redis.SaveLink(ctx, l.UserID, generatedLink, l.OriginalURL, int32(l.ID))
	if err != nil {
		s.logger.Error("Failed to save link to Redis",
			zap.Int64("user", req.UserId),
			zap.String("desc", req.Description),
			zap.Int32("url_id", req.UrlId),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Internal, "Failed to save link to cache: %v", err)
	}

	s.logger.Debug("Saved to redis",
		zap.Int64("user", req.UserId),
		zap.String("desc", req.Description),
		zap.String("gen_url", generatedLink),
	)

	fullURL := getFullLink(s.cfg.BaseURL, l.UserID, generatedLink)

	s.logger.Debug("Generated Full Link",
		zap.String("full_link", fullURL),
	)

	return &gen.GetLinkResponse{GeneratedUrl: fullURL}, nil
}

func (s *LinkService) GetAllLinks(ctx context.Context, req *gen.GetAllLinksRequest) (*gen.GetAllLinksResponse, error) {
	s.logger.Info("New req GetAllLinks()",
		zap.Int64("user", req.UserId),
	)

	links, err := s.db.GetUserLinks(ctx, req.UserId)
	if err != nil {
		s.logger.Error("Failed to get all links from DB",
			zap.Int64("user", req.UserId),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Aborted, "Can't get all links from DB, err: %s", err)
	}

	resp := &gen.GetAllLinksResponse{
		Links: links,
	}

	return resp, nil
}
