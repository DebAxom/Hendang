# Hendang

**Hendang** is a CLI tool that allows you to split large files into smaller chunks, store them in a directory (with `.chd` suffix), and merge them back into the original file.  

It also supports downloading chunked files from the internet — making it resilient against connection failures. If a download is interrupted, you don’t need to start over; only the missing chunks are fetched.

## Features
- Split large files into smaller, manageable `.chunk` parts.
- Merge `.chunk` files back into the original file.
- Download files in chunks from the internet.
- Resume downloads — only missing chunks are redownloaded.  
- Cross-platform support (Linux, macOS, Windows).

## Installation
```bash
git clone https://github.com/DebAxom/Hendang.git
cd Hendang
go build -o hendang
```
Add `hendang` to path (different for each OS)

## Usage

### Breake a file into chunks
The `break` command breaks the file into chunks and creates a new folder with `.chd` extension. You can set the output directory name, but it's optional.
```
hendang cut <filename> <outputdir(optional)>
```

### Merging the chunks back to original file
The `merge` command takes 2 arguments
```
hendang merge <dirname.chd> <outputfile>
```

### Download from net
```
hendang download <url> <filename>
```

### Deleting partially-downloaded chunks
```
hendang reset
```

## Downloading a test file
The test file is a Thai Ganesh prayer song.
```
hendang download https://raw.githubusercontent.com/DebAxom/Hendang/refs/heads/main/download/thai-ganesh.chd thai-ganesh-prayer
```
