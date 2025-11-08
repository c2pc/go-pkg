package file_manager

import "errors"

var ErrObjectIsNotExists = errors.New("dir is not exists")
var ErrObjectIsNotDirectory = errors.New("object is not directory")
var ErrObjectIsNotFile = errors.New("object is not file")
var ErrObjectIsAlreadyExists = errors.New("object is already exists")
var ErrObjectHasInvalidExtension = errors.New("object has invalid extension")
var ErrObjectIsNotWritable = errors.New("object is not writable")
