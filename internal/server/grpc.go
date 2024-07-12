package server

import (
	// импортируем пакет со сгенерированными protobuf-файлами
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/winkor4/taktaev-yandex-dev-uri.git/internal/model"
	pb "github.com/winkor4/taktaev-yandex-dev-uri.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Команды для создания proto
// protoc --go_out=. --go_opt=paths=source_relative \
//   --go-grpc_out=. --go-grpc_opt=paths=source_relative \
//   proto/demo.proto

// Создаем пользователя
func (s *ServerGRPC) authorizationInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var user string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("auth_token")
		if len(values) > 0 {
			user = values[0]
		}
	}
	if len(user) == 0 {
		user = uuid.New().String()
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "auth_token", user)
	s.user = user
	return handler(ctx, req)
}

// Сокращение ссылки
func (s *ServerGRPC) ShortURL(ctx context.Context, in *pb.ShortURLRequest) (*pb.ShortURLResponse, error) {
	var response pb.ShortURLResponse

	ourl := in.Url
	user := s.user

	urls := make([]model.URL, 1)
	urls[0].Key = model.ShortKey(ourl)
	urls[0].OriginalURL = ourl
	urls[0].UserID = user

	err := s.urlRepo.SaveURL(urls)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	result := fmt.Sprintf(s.cfg.ResSrvAdr+"/%s", urls[0].Key)

	resUser := new(pb.User)
	resUser.Name = user

	response.ShortenUrl = result
	response.User = resUser

	return &response, nil
}

// Получить ссылку
func (s *ServerGRPC) GetURL(ctx context.Context, in *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	var response pb.GetURLResponse

	key := in.Id
	url, err := s.urlRepo.GetURL(key)
	if err == model.ErrIsDeleted {
		return nil, status.Errorf(codes.Canceled, err.Error())
	}
	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	response.Url = url.OriginalURL

	return &response, nil
}

// Сократить список ссылок
func (s *ServerGRPC) ShortBatch(ctx context.Context, in *pb.ShortBatchRequest) (*pb.ShortBatchResponse, error) {
	var response pb.ShortBatchResponse

	user := in.User.Name

	urls := in.Urls

	data := make([]model.URL, len(urls))

	for i, ourl := range urls {
		data[i].Key = model.ShortKey(ourl)
		data[i].OriginalURL = ourl
		data[i].UserID = user

		response.ShortenUrls = append(response.ShortenUrls, fmt.Sprintf(s.cfg.ResSrvAdr+"/%s", data[i].Key))
	}

	err := s.urlRepo.SaveURL(data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &response, nil
}

// Удалить список ссылок
func (s *ServerGRPC) DeleteURL(ctx context.Context, in *pb.DeleteURLRequest) (*pb.DeleteURLResponse, error) {
	var response pb.DeleteURLResponse

	user := in.User.Name
	keys := in.Keys

	var data delURL
	data.user = user
	data.keys = keys
	go putDelURL(s.deleteCh, data)

	return &response, nil
}
