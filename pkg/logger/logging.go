package logger

import "github.com/sirupsen/logrus"

// Field log object
type Field struct {
	Key   string
	Value interface{}
}

// FieldFunc log
type FieldFunc func(key string, value interface{}) *Field

type Fields []Field

// NewFields create instance new field
func NewFields(p ...Field) Fields {
	x := Fields{}

	for i := 0; i < len(p); i++ {
		x.Append(p[i])
	}

	return x
}

// Append new field
func (f *Fields) Append(p Field) {
	*f = append(*f, p)
}

// Any log
func Any(k string, v interface{}) Field {
	return Field{
		Key:   k,
		Value: v,
	}
}

// String log
func String(k string, v string) Field {
	return Field{
		Key:   k,
		Value: v,
	}
}

// EventName log
func EventName(v interface{}) Field {
	return Field{
		Key:   EventNameKey,
		Value: v,
	}
}

func extract(args ...Field) map[string]interface{} {
	if len(args) == 0 {
		return nil
	}

	data := map[string]interface{}{}
	for _, fl := range args {
		data[fl.Key] = fl.Value
	}
	return data
}

// Error log
func Error(arg interface{}, fl ...Field) {
	logrus.WithFields(
		addField(logrus.Fields{
			EventKey: extract(fl...),
		}),
	).Error(arg)

}

func Info(arg interface{}, fl ...Field) {
	logrus.WithFields(
		addField(logrus.Fields{
			EventKey: extract(fl...),
		}),
	).Info(arg)
}

func Debug(arg interface{}, fl ...Field) {
	logrus.WithFields(
		addField(logrus.Fields{
			EventKey: extract(fl...),
		}),
	).Debug(arg)
}

// Fatal log
func Fatal(arg interface{}, fl ...Field) {
	logrus.WithFields(
		addField(logrus.Fields{
			EventKey: extract(fl...),
		}),
	).Fatal(arg)
}

// Warn log
func Warn(arg interface{}, fl ...Field) {
	logrus.WithFields(
		addField(logrus.Fields{
			EventKey: extract(fl...),
		}),
	).Warn(arg)
}

// Trace log
func Trace(arg interface{}, fl ...Field) {
	logrus.WithFields(
		addField(logrus.Fields{
			EventKey: extract(fl...),
		}),
	).Trace(arg)
}

func addField(f logrus.Fields) logrus.Fields {
	return f
}
