async function compressData(data, encoding = GZIP_MARKER, callback) {
    console.log("compressing with", encoding)
    if (encoding == GZIP_MARKER) {
      return import("/js/gzip/pako.js").then((module) => {
        console.log({gzdata:data})
        return pako.deflate(data, {level:"9"});
      });
    } else if (encoding == BROT_MARKER) {
      
    } else if (encoding == LZMA_MARKER) {
      return new Promise(function(resolve, reject) {
        console.log({xz:data})
  
        LZMA.compress(data, 9, function(result, error) {
          if (error) reject(error);
          resolve(result);
        });
      });
    } 
  }