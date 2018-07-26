package store

import (
	"github.com/golang/glog"
	"github.com/tinylib/msgp/msgp"

	"github.com/hashicorp/raft"
)

type persister func(*Messages, *msgp.Writer, raft.SnapshotSink) error

type fsmSnapShot struct {
	messages   *Messages
	persisters []persister
}

func (f *fsmSnapShot) Release() {
	glog.Info("release =>")
}

func (f *fsmSnapShot) Persist(sink raft.SnapshotSink) error {
	glog.Info("persist =>")

	msgpWriter := msgp.NewWriter(sink)

	for _, fn := range f.persisters {
		if err := fn(f.messages, msgpWriter, sink); err != nil {
			err = sink.Cancel()
			glog.Errorf("persist err %v\n", err)
			return err
		}
	}

	return nil
}

func persistRules(messages *Messages, writer *msgp.Writer, sink raft.SnapshotSink) error {

	for _, rule := range messages.Rules {
		if _, err := sink.Write([]byte{byte(RuleType)}); err != nil {
			glog.Errorf("persistRules %v", err)
			continue
		}

		glog.Info("persist rule msg size ", rule.Msgsize())
		// Encode message.
		err := rule.EncodeMsg(writer)
		if err != nil {
			glog.Errorf("persistRules %v", err)
			continue
		}

		err = writer.Flush()
		glog.Infof("persistRules %+v %v\n", rule, err)

	}

	return nil
}

func persistScripts(messages *Messages, writer *msgp.Writer, sink raft.SnapshotSink) error {

	for _, script := range messages.Scripts {
		if _, err := sink.Write([]byte{byte(ScriptType)}); err != nil {
			glog.Errorf("persistScripts %v", err)
			continue
		}

		glog.Info("persist script msg size ", script.Msgsize())

		// Encode message.
		err := script.EncodeMsg(writer)
		if err != nil {
			glog.Errorf("persistScripts %v", err)
			continue
		}

		err = writer.Flush()
		glog.Infof("persistScripts %+v %v \n", script, err)
	}
	return nil
}

func persistRecords(messages *Messages, writer *msgp.Writer, sink raft.SnapshotSink) error {

	for _, record := range messages.Records {
		if _, err := sink.Write([]byte{byte(RecordType)}); err != nil {
			glog.Errorf("persistRecords %v", err)
			continue
		}

		glog.Info("persist record msg size ", record.Msgsize())
		// Encode message.
		err := record.EncodeMsg(writer)
		if err != nil {
			glog.Errorf("persistRecords %v", err)
			continue
		}

		err = writer.Flush()
		glog.Infof("persistRecords %+v %v \n", record, err)
	}
	return nil
}
