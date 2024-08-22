package service

import (
	"context"

	"github.com/0x0FACED/link-saver-api/internal/domain/models"
	"github.com/0x0FACED/proto-files/link_service/gen"
	"github.com/gocolly/colly"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *LinkService) SaveLink(ctx context.Context, req *gen.SaveLinkRequest) (*gen.SaveLinkResponse, error) {
	s.Logger.Debug("Received link",
		zap.Int64("user", req.UserId),
		zap.String("desc", req.Description),
		zap.String("url", req.OriginalUrl),
	)

	// onHTML -> html page, save to page.html
	s.colly.OnHTML("html", func(e *colly.HTMLElement) {
		link := &models.Link{
			OriginalURL: req.OriginalUrl,
			UserID:      req.UserId,
			Description: req.Description,
			Content:     []byte(e.Response.Body),
		}
		s.Logger.Debug("Visited link", zap.String("url", link.OriginalURL))
		// save page as bytea to database
		err := s.saveToDatabase(context.TODO(), link)
		if err != nil {
			s.Logger.Error("Failed to save to db: " + err.Error())
			return
		}
		s.Logger.Debug("Link successfully saved to db")
	})

	// onHTML -> image, save to imgURL.png/jpg file
	/*c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		imgURL := resolveURL(e.Attr("src"), e.Request.URL)
		err := downloadFile(outputDir, imgURL)
		if err != nil {
			log.Printf("Error saving image: %v", err)
		}
	})*/

	// onHTML -> stylesheet (CSS), save to css file
	/*c.OnHTML("link[rel='stylesheet']", func(e *colly.HTMLElement) {
		cssURL := resolveURL(e.Attr("href"), e.Request.URL)
		err := downloadFile(outputDir, cssURL)
		if err != nil {
			log.Printf("Error saving CSS: %v", err)
		}
	})*/

	// onHTML -> script (js), save to js file
	/*c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		jsURL := resolveURL(e.Attr("src"), e.Request.URL)
		err := downloadFile(outputDir, jsURL)
		if err != nil {
			log.Printf("Error saving JavaScript: %v", err)
		}
	})*/

	// start parse html page
	err := s.colly.Visit(req.OriginalUrl)
	if err != nil {
		s.Logger.Error("Error while scrap HTML",
			zap.String("err", err.Error()),
			zap.Int64("user", req.UserId),
			zap.String("desc", req.Description),
			zap.String("url", req.OriginalUrl),
		)
		return &gen.SaveLinkResponse{Success: false}, err
	}

	s.Logger.Info("Finished", zap.Int64("user", req.UserId))
	return &gen.SaveLinkResponse{Success: true}, nil
}

func (s *LinkService) DeleteLink(ctx context.Context, req *gen.DeleteLinkRequest) (*gen.DeleteLinkResponse, error) {
	// TODO: impl
	return &gen.DeleteLinkResponse{Success: true}, nil
}

func (s *LinkService) GetLinks(ctx context.Context, req *gen.GetLinksRequest) (*gen.GetLinksResponse, error) {
	s.Logger.Debug("New req GetLinks()",
		zap.Int64("user", req.UserId),
		zap.String("desc", req.Description),
	)

	links, err := s.db.GetLinksByTelegramIDDesc(ctx, req.UserId, req.Description)
	if err != nil {
		s.Logger.Error("Failed to get links by username and desc",
			zap.Int64("user", req.UserId),
			zap.String("desc", req.Description),
		)
		return &gen.GetLinksResponse{Links: nil}, err
	}
	s.Logger.Debug("Found links", zap.Int64("user", req.UserId), zap.Any("links", links))
	return &gen.GetLinksResponse{Links: links}, nil
}

func (s *LinkService) GetLink(ctx context.Context, req *gen.GetLinkRequest) (*gen.GetLinkResponse, error) {
	s.Logger.Debug("New req GetLink()",
		zap.Int64("user", req.UserId),
		zap.String("desc", req.Description),
		zap.Int32("url_id", req.UrlId),
	)

	l, err := s.db.GetLinkByID(ctx, int(req.UrlId))
	if err != nil {
		s.Logger.Error("Failed to get link by id from Postgres",
			zap.Int64("user", req.UserId),
			zap.String("desc", req.Description),
			zap.Error(err),
		)
		return nil, err
	}

	redisLink, err := s.redis.GetLink(ctx, req.UserId, l.OriginalURL)
	if err != nil && err != redis.Nil {
		s.Logger.Error("Failed to get link from Redis",
			zap.Int64("user", req.UserId),
			zap.String("desc", req.Description),
			zap.Error(err),
		)
		return nil, err
	}

	if redisLink != nil {
		s.Logger.Debug("Link found in Redis",
			zap.Int64("user", req.UserId),
			zap.String("desc", req.Description),
		)

		fullURL := getFullLink(l.UserID, redisLink.Link)
		s.Logger.Debug("Generated Full Link",
			zap.String("full_link", fullURL),
		)
		return &gen.GetLinkResponse{GeneratedUrl: fullURL}, nil
	}

	generatedLink := hash(l.UserID, l.OriginalURL)

	err = s.redis.SaveLink(ctx, l.UserID, generatedLink, l.OriginalURL, int32(l.ID))
	if err != nil {
		s.Logger.Error("Failed to save link to Redis",
			zap.Int64("user", req.UserId),
			zap.String("desc", req.Description),
			zap.Int32("url_id", req.UrlId),
			zap.Error(err),
		)
		return nil, err
	}

	s.Logger.Debug("Saved to redis",
		zap.Int64("user", req.UserId),
		zap.String("desc", req.Description),
		zap.String("gen_url", generatedLink),
	)

	fullURL := getFullLink(l.UserID, generatedLink)

	s.Logger.Debug("Generated Full Link",
		zap.String("full_link", fullURL),
	)

	return &gen.GetLinkResponse{GeneratedUrl: fullURL}, nil
}

func (s *LinkService) GetAllLinks(ctx context.Context, req *gen.GetAllLinksRequest) (*gen.GetAllLinksResponse, error) {
	s.Logger.Debug("New req GetAllLinks()",
		zap.Int64("user", req.UserId),
	)

	links, err := s.db.GetLinksByTelegramIDDesc(ctx, req.UserId, "")
	if err != nil {
		s.Logger.Error("Failed to get all links from DB",
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
