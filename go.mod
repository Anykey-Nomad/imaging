module github.com/disintegration/imaging

go 1.25.0

require golang.org/x/image v0.0.0-20191009234506-e7c1f5e7dbb8

require (
	github.com/go-webgpu/goffi v0.5.5 // indirect
	github.com/go-webgpu/webgpu v0.5.2 // indirect
	github.com/gogpu/gpucontext v0.21.0 // indirect
	github.com/gogpu/gputypes v0.5.0 // indirect
	github.com/gogpu/naga v0.17.15 // indirect
	github.com/gogpu/wgpu v0.30.4 // indirect
	golang.org/x/sys v0.46.0 // indirect
)

replace golang.org/x/image => ../image-gogpu
