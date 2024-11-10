package sync

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/c2pc/go-musthave-metrics/internal/storage"
)

const (
	separator = string("\t")
)

type Storager interface {
	GetName() string
	GetString(ctx context.Context, key string) (string, error)
	GetAllString(ctx context.Context) (map[string]string, error)
	SetString(ctx context.Context, values ...storage.Valuer[string]) error
}

type Sync struct {
	file     *os.File
	storages map[string]Storager
}

type Config struct {
	StoreInterval   int64
	FileStoragePath string
	Restore         bool
}

func Start(ctx context.Context, cfg Config, storages ...Storager) (io.Closer, error) {
	if len(storages) == 0 {
		return nil, errors.New("no storages")
	}

	if cfg.StoreInterval < 0 {
		return nil, errors.New("invalid store interval")
	}

	var storagesMap = make(map[string]Storager, len(storages))
	for _, storager := range storages {
		storagesMap[storager.GetName()] = storager
	}

	s := &Sync{
		file:     nil,
		storages: storagesMap,
	}

	err := s.openFile(cfg.FileStoragePath)
	if err != nil {
		return nil, err
	}

	if cfg.Restore {
		if err := s.restore(ctx); err != nil {
			return nil, err
		}
	}

	if cfg.StoreInterval > 0 {
		go s.listen(ctx, time.Duration(cfg.StoreInterval)*time.Second)
	}

	return s, nil
}

func (s *Sync) Close() error {
	return s.file.Close()
}

func (s *Sync) listen(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := s.saveDataToFile(ctx)
			if err != nil {
				log.Printf("failed to save data to file: %v", err)
			}
		}
	}
}

func (s *Sync) restore(ctx context.Context) error {
	if s.file == nil {
		return errors.New("file not open")
	}

	_, err := s.file.Seek(0, 0)
	if err != nil {
		return err
	}

	var dataList []column
	scanner := bufio.NewScanner(s.file)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" || line == "\n" {
			continue
		}

		d, err := s.lineToData(line)
		if err != nil {
			return err
		}
		dataList = append(dataList, *d)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	err = s.writeDataToStorage(ctx, dataList...)
	if err != nil {
		return err
	}

	return nil
}

func (s *Sync) saveDataToFile(ctx context.Context) error {
	data, err := s.readDataFromStorage(ctx)
	if err != nil {
		return err
	}

	var text string
	for _, d := range data {
		text = text + s.dataToLine(d) + "\n"
	}

	err = s.writeToFile(text)
	if err != nil {
		return err
	}

	return nil
}

type column struct {
	name  string
	key   string
	value string
}

func (s *Sync) lineToData(line string) (*column, error) {
	split := strings.Split(line, separator)
	if len(split) != 3 {
		return nil, errors.New("invalid line")
	}

	return &column{split[0], split[1], split[2]}, nil
}

func (s *Sync) dataToLine(data column) string {
	return fmt.Sprintf("%s%s%s%s%s", data.name, separator, data.key, separator, data.value)
}

func (s *Sync) readDataFromStorage(ctx context.Context) ([]column, error) {
	var dataList []column
	for _, storager := range s.storages {
		data, err := storager.GetAllString(ctx)
		if err != nil {
			return nil, err
		}
		for k, v := range data {
			dataList = append(dataList, column{storager.GetName(), k, v})
		}
	}
	return dataList, nil
}

func (s *Sync) writeDataToStorage(ctx context.Context, data ...column) error {
	columns := map[string][]storage.Valuer[string]{}

	for _, d := range data {
		if _, ok := s.storages[d.name]; !ok {
			return fmt.Errorf("storage %s not found", d.name)
		}

		if _, ok := columns[d.name]; !ok {
			columns[d.name] = []storage.Valuer[string]{}
		}

		columns[d.name] = append(columns[d.name], storage.Value[string]{Key: d.key, Value: d.value})
	}

	for key, value := range columns {
		err := s.storages[key].SetString(ctx, value...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Sync) openFile(filePath string) error {
	dir := filepath.Dir(filePath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	s.file = file

	return nil
}

func (s *Sync) writeToFile(data string) error {
	if s.file == nil {
		return errors.New("file not open")
	}

	err := s.file.Truncate(0)
	if err != nil {
		return err
	}

	_, err = s.file.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = s.file.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}
