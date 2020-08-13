package log

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"net"
)

var pool = buffer.NewPool()

type protobufEncoder struct {
	zapcore.Encoder
	serviceName string
}

func newProtobufEncoder(cfg zapcore.EncoderConfig, serviceName string) zapcore.Encoder {
	return protobufEncoder{
		Encoder:     zapcore.NewJSONEncoder(cfg),
		serviceName: serviceName,
	}
}

func (p protobufEncoder) Clone() zapcore.Encoder {
	return protobufEncoder{Encoder: p.Encoder.Clone()}
}

func (p protobufEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	le := LoggingEvent{
		ServiceName:     p.serviceName,
		Message:         ent.Message,
		TimeStamp:       uint64(ent.Time.UnixNano() / 1000 / 1000),
		LoggerName:      ent.LoggerName,
		ErrorTraceStack: ent.Stack,
	}
	if ent.Caller.Defined {
		le.CallerFileName = ent.Caller.TrimmedPath()
		le.CallerLineIndex = int32(ent.Caller.Line)
	}
	switch ent.Level {
	case zapcore.DebugLevel:
		le.Level = LoggingEvent_DEBUG
	case zapcore.InfoLevel:
		le.Level = LoggingEvent_INFO
	case zapcore.WarnLevel:
		le.Level = LoggingEvent_WARN
	case zapcore.ErrorLevel:
		fallthrough
	case zapcore.DPanicLevel:
		fallthrough
	case zapcore.PanicLevel:
		fallthrough
	case zapcore.FatalLevel:
		le.Level = LoggingEvent_ERROR
	}

	data, err := proto.Marshal(&le)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	buf := pool.Get()

	// 写入长度
	lend := proto.EncodeVarint(uint64(len(data)))
	_, err = buf.Write(lend)
	if err != nil {
		buf.Free()
		return nil, errors.WithStack(err)
	}

	// 写入数据
	_, err = buf.Write(data)
	if err != nil {
		buf.Free()
		return nil, errors.WithStack(err)
	}

	return buf, nil
}

type protobufWriterSyncer struct {
	conn net.Conn
}

func newProtobufWriterSyncer(address string) (zapcore.WriteSyncer, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
	}
	return &protobufWriterSyncer{conn: conn}, nil
}

func (s *protobufWriterSyncer) Write(bs []byte) (int, error) {
	return s.conn.Write(bs)
}

func (s *protobufWriterSyncer) Sync() error {
	return s.conn.Close()
}
