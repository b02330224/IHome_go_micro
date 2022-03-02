package fdfs_client

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	logger                                          = logrus.New()
	storagePoolChan      chan *storagePool          = make(chan *storagePool, 1)
	storagePoolMap       map[string]*ConnectionPool = make(map[string]*ConnectionPool)
	fetchStoragePoolChan chan interface{}           = make(chan interface{}, 1)
	quit                 chan bool
)

type FdfsClient struct {
	tracker     *Tracker
	trackerPool *ConnectionPool
	timeout     int
}

type Tracker struct {
	HostList []string
	Port     int
}

type storagePool struct {
	storagePoolKey string
	hosts          []string
	port           int
	minConns       int
	maxConns       int
}

func init() {
	logger.Formatter = new(logrus.TextFormatter)
	runtime.GOMAXPROCS(runtime.NumCPU())
	go func() {
		// start a loop
		for {
			select {
			case spd := <-storagePoolChan:
				if sp, ok := storagePoolMap[spd.storagePoolKey]; ok {
					fetchStoragePoolChan <- sp
				} else {
					var (
						sp  *ConnectionPool
						err error
					)
					sp, err = NewConnectionPool(spd.hosts, spd.port, spd.minConns, spd.maxConns)
					if err != nil {
						fetchStoragePoolChan <- err
					} else {
						storagePoolMap[spd.storagePoolKey] = sp
						fetchStoragePoolChan <- sp
					}
				}
			case <-quit:
				break
			}
		}
	}()
}

func getTrackerConf(confPath string) (*Tracker, error) {
	fc := &FdfsConfigParser{}
	cf, err := fc.Read(confPath)
	if err != nil {
		logger.Errorf("Read conf error :%s", err)
		return nil, err
	}
	trackerListString, _ := cf.RawString("DEFAULT", "tracker_server")
	trackerList := strings.Split(trackerListString, ",")

	var (
		trackerIpList []string
		trackerPort   string = "22122"
	)

	for _, tr := range trackerList {
		var trackerIp string
		tr = strings.TrimSpace(tr)
		parts := strings.Split(tr, ":")
		trackerIp = parts[0]
		if len(parts) == 2 {
			trackerPort = parts[1]
		}
		if trackerIp != "" {
			trackerIpList = append(trackerIpList, trackerIp)
		}
	}
	tp, err := strconv.Atoi(trackerPort)
	tracer := &Tracker{
		HostList: trackerIpList,
		Port:     tp,
	}
	return tracer, nil
}

func NewFdfsClient(confPath string) (*FdfsClient, error) {
	tracker, err := getTrackerConf(confPath)
	if err != nil {
		return nil, err
	}

	trackerPool, err := NewConnectionPool(tracker.HostList, tracker.Port, 10, 150)
	if err != nil {
		return nil, err
	}

	return &FdfsClient{tracker: tracker, trackerPool: trackerPool}, nil
}

func NewFdfsClientByTracker(tracker *Tracker) (*FdfsClient, error) {
	trackerPool, err := NewConnectionPool(tracker.HostList, tracker.Port, 10, 150)
	if err != nil {
		return nil, err
	}

	return &FdfsClient{tracker: tracker, trackerPool: trackerPool}, nil
}

func CloseFdfsClient() {
	quit <- true
}

func (this *FdfsClient) UploadByFilename(filename string) (*UploadFileResponse, error) {
	if err := fdfsCheckFile(filename); err != nil {
		return nil, errors.New(err.Error() + "(uploading)")
	}

	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageStorWithoutGroup()
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageUploadByFilename(tc, storeServ, filename)
}

func (this *FdfsClient) UploadByBuffer(filebuffer []byte, fileExtName string) (*UploadFileResponse, error) {
	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageStorWithoutGroup()
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageUploadByBuffer(tc, storeServ, filebuffer, fileExtName)
}

func (this *FdfsClient) UploadSlaveByFilename(filename, remoteFileId, prefixName string) (*UploadFileResponse, error) {
	if err := fdfsCheckFile(filename); err != nil {
		return nil, errors.New(err.Error() + "(uploading)")
	}

	tmp, err := splitRemoteFileId(remoteFileId)
	if err != nil || len(tmp) != 2 {
		return nil, err
	}
	groupName := tmp[0]
	remoteFilename := tmp[1]

	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageStorWithGroup(groupName)
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageUploadSlaveByFilename(tc, storeServ, filename, prefixName, remoteFilename)
}

func (this *FdfsClient) UploadSlaveByBuffer(filebuffer []byte, remoteFileId, fileExtName string) (*UploadFileResponse, error) {
	tmp, err := splitRemoteFileId(remoteFileId)
	if err != nil || len(tmp) != 2 {
		return nil, err
	}
	groupName := tmp[0]
	remoteFilename := tmp[1]

	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageStorWithGroup(groupName)
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageUploadSlaveByBuffer(tc, storeServ, filebuffer, remoteFilename, fileExtName)
}

func (this *FdfsClient) UploadAppenderByFilename(filename string) (*UploadFileResponse, error) {
	if err := fdfsCheckFile(filename); err != nil {
		return nil, errors.New(err.Error() + "(uploading)")
	}

	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageStorWithoutGroup()
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageUploadAppenderByFilename(tc, storeServ, filename)
}

func (this *FdfsClient) UploadAppenderByBuffer(filebuffer []byte, fileExtName string) (*UploadFileResponse, error) {
	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageStorWithoutGroup()
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageUploadAppenderByBuffer(tc, storeServ, filebuffer, fileExtName)
}

func (this *FdfsClient) DeleteFile(remoteFileId string) error {
	tmp, err := splitRemoteFileId(remoteFileId)
	if err != nil || len(tmp) != 2 {
		return err
	}
	groupName := tmp[0]
	remoteFilename := tmp[1]

	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageUpdate(groupName, remoteFilename)
	if err != nil {
		return err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageDeleteFile(tc, storeServ, remoteFilename)
}

func (this *FdfsClient) DownloadToFile(localFilename string, remoteFileId string, offset int64, downloadSize int64) (*DownloadFileResponse, error) {
	tmp, err := splitRemoteFileId(remoteFileId)
	if err != nil || len(tmp) != 2 {
		return nil, err
	}
	groupName := tmp[0]
	remoteFilename := tmp[1]

	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageFetch(groupName, remoteFilename)
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageDownloadToFile(tc, storeServ, localFilename, offset, downloadSize, remoteFilename)
}

func (this *FdfsClient) DownloadToBuffer(remoteFileId string, offset int64, downloadSize int64) (*DownloadFileResponse, error) {
	tmp, err := splitRemoteFileId(remoteFileId)
	if err != nil || len(tmp) != 2 {
		return nil, err
	}
	groupName := tmp[0]
	remoteFilename := tmp[1]

	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageFetch(groupName, remoteFilename)
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	var fileBuffer []byte
	return store.storageDownloadToBuffer(tc, storeServ, fileBuffer, offset, downloadSize, remoteFilename)
}

func (this *FdfsClient) getStoragePool(ipAddr string, port int) (*ConnectionPool, error) {
	hosts := []string{ipAddr}
	var (
		storagePoolKey string = fmt.Sprintf("%s-%d", hosts[0], port)
		result         interface{}
		err            error
		ok             bool
	)

	spd := &storagePool{
		storagePoolKey: storagePoolKey,
		hosts:          hosts,
		port:           port,
		minConns:       10,
		maxConns:       150,
	}
	storagePoolChan <- spd
	for {
		select {
		case result = <-fetchStoragePoolChan:
			var storagePool *ConnectionPool
			if err, ok = result.(error); ok {
				return nil, err
			} else if storagePool, ok = result.(*ConnectionPool); ok {
				return storagePool, nil
			} else {
				return nil, errors.New("none")
			}
		}
	}
}
