package service

import (
	"context"
	"net/http"
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
	statusCode := -1
	var link *models.Link

	// JS обработчик
	s.colly.OnHTML("script[src]", func(e *colly.HTMLElement) {
		s.logger.Info("Visited script", zap.String("text", e.Attr("src")))
		jsPath := e.Attr("src")
		var jsURL string

		if strings.HasPrefix(jsPath, "/") {
			jsURL = e.Request.AbsoluteURL(jsPath)
		} else {
			jsURL = e.Request.URL.JoinPath(jsPath).String() // Преобразование URL в строку
		}

		relPath := getRelativePath(jsURL)
		s.logger.Debug("Relative Path", zap.String("path", relPath))

		jsContent := s.fetchResourceContent(jsURL, e)
		if jsContent != nil {
			res := &models.Resource{
				Name:    hashResName(relPath, models.ScriptType),
				Content: jsContent,
				Type:    models.ScriptType,
			}
			err := s.saveResource(res)
			if err != nil {
				s.logger.Debug("Failed to save script", zap.Any("script", res.Name))
			}
			newPath := getResourceURL(s.cfg.BaseURL, "script", res.Name)
			e.DOM.SetAttr("src", newPath)
			attr, ex := e.DOM.Attr("src")
			if ex {
				s.logger.Info("Attr", zap.String("attr", attr))
			}
		}
	})

	// CSS обработчик
	s.colly.OnHTML("link[rel='stylesheet']", func(e *colly.HTMLElement) {
		s.logger.Info("Visited style", zap.String("text", e.Attr("href")))
		cssPath := e.Attr("href")
		var cssURL string

		if strings.HasPrefix(cssPath, "/") {
			cssURL = e.Request.AbsoluteURL(cssPath)
		} else {
			cssURL = e.Request.URL.JoinPath(cssPath).String() // Преобразование URL в строку
		}

		relPath := getRelativePath(cssURL)
		s.logger.Debug("Relative Path", zap.String("path", relPath))

		cssContent := s.fetchResourceContent(cssURL, e)
		if cssContent != nil {
			res := &models.Resource{
				Name:    hashResName(relPath, models.CSSType),
				Content: cssContent,
				Type:    models.CSSType,
			}
			err := s.saveResource(res)
			if err != nil {
				s.logger.Debug("Failed to save CSS", zap.Any("css", res.Name))
			}
			newPath := getResourceURL(s.cfg.BaseURL, "css", res.Name)
			e.DOM.SetAttr("href", newPath)
			attr, ex := e.DOM.Attr("href")
			if ex {
				s.logger.Info("Attr", zap.String("attr", attr))
			}
		}
	})

	// Обработчик изображений
	s.colly.OnHTML("img[src]", func(e *colly.HTMLElement) {
		s.logger.Info("Visited image", zap.String("text", e.Attr("src")))
		imgPath := e.Attr("src")

		var imgURL string
		if strings.HasPrefix(imgPath, "/") {
			imgURL = e.Request.AbsoluteURL(imgPath)
		} else {
			imgURL = e.Request.URL.JoinPath(imgPath).String()
		}

		relPath := getRelativePath(imgURL)
		s.logger.Debug("Relative Path", zap.String("path", relPath))

		imgContent := s.fetchResourceContent(imgURL, e)
		if imgContent != nil {
			res := &models.Resource{
				Name:    hashResName(relPath, models.ImageType),
				Content: imgContent,
				Type:    models.ImageType,
			}
			err := s.saveResource(res)
			if err != nil {
				s.logger.Debug("Failed to save image", zap.Any("image", res.Name))
			}
			newPath := getResourceURL(s.cfg.BaseURL, "image", res.Name)
			e.DOM.SetAttr("src", newPath)
			attr, ex := e.DOM.Attr("src")
			if ex {
				s.logger.Info("Attr", zap.String("attr", attr))
			}
		}
	})

	// Обработчик для HTML
	s.colly.OnHTML("html", func(e *colly.HTMLElement) {
		// Присваиваем HTML-страницу только после обработки всех ресурсов
		updatedHTML, err := e.DOM.Html()
		if err != nil {
			s.logger.Error("Error DOM.Html()", zap.Error(err))
		}
		link = &models.Link{
			OriginalURL: req.OriginalUrl,
			UserID:      req.UserId,
			Description: req.Description,
			Content:     []byte(updatedHTML),
		}
		s.logger.Info("Visited link", zap.String("url", link.OriginalURL))
	})

	// onError -> to handle error in scrap
	s.colly.OnError(func(r *colly.Response, e error) {
		statusCode = r.StatusCode
		s.logger.Debug("OnError()", zap.Error(e), zap.Int("status_code", r.StatusCode))
	})

	// start colly
	err := s.colly.Visit(req.OriginalUrl)
	s.colly.Wait()
	if err != nil {
		s.logger.Error("Error while scrap HTML",
			zap.Error(err),
			zap.Int64("user", req.UserId),
			zap.String("desc", req.Description),
			zap.String("url", req.OriginalUrl),
		)
		return &gen.SaveLinkResponse{Success: false, Message: "Not Saved, invalid link"}, status.Errorf(codes.InvalidArgument, "Invalid link: %v", err)
	}
	if statusCode == http.StatusForbidden {
		s.logger.Debug("Not Saved", zap.Int64("user", req.UserId))
		return &gen.SaveLinkResponse{Success: false, Message: "Not Saved, invalid link"}, status.Errorf(codes.InvalidArgument, "You dont have permissions")
	}

	if link == nil {
		s.logger.Debug("Not Saved", zap.Int64("user", req.UserId))
		return &gen.SaveLinkResponse{Success: false, Message: "Not Saved, invalid link"}, status.Errorf(codes.InvalidArgument, "Link data is missing")
	}

	s.logger.Info("Finished", zap.Int64("user", req.UserId))

	// save page as bytea to database
	err = s.saveToDatabase(context.TODO(), link)
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
