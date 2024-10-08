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
	s.logger.Debug("Received link",
		zap.Int64("user", req.UserId),
		zap.String("desc", req.Description),
		zap.String("url", req.OriginalUrl),
	)
	statusCode := -1
	var link *models.Link
	// onHTML -> html page visiting
	s.colly.OnHTML("html", func(e *colly.HTMLElement) {
		link = &models.Link{
			OriginalURL: req.OriginalUrl,
			UserID:      req.UserId,
			Description: req.Description,
			Content:     []byte(e.Response.Body),
		}
		s.logger.Debug("Visited link", zap.String("url", link.OriginalURL))
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

	if statusCode == 0 || link == nil {
		s.logger.Info("Not Saved", zap.Int64("user", req.UserId))
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

	return &gen.SaveLinkResponse{Success: true, Message: "Succeefully saved"}, nil
}

func (s *LinkService) DeleteLink(ctx context.Context, req *gen.DeleteLinkRequest) (*gen.DeleteLinkResponse, error) {
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
	s.logger.Debug("New req GetLinks()",
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
	s.logger.Debug("New req GetLink()",
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
	s.logger.Debug("New req GetAllLinks()",
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
