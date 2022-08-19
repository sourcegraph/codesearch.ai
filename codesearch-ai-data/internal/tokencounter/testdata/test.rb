def gzip(string)
    result = StringIO.new
    zio = Zlib::GzipWriter.new(result, nil, nil)
    zio.mtime = 1
    zio.write(string)
    zio.close
    result.string
end

def open_zip(filename)
    @zipfile = Zip::File.open(filename)
    @clist.clear
    @zipfile.each do |entry|
        @clist.append([entry.name,
                        entry.size.to_s,
                        (100.0 * entry.compressedSize / entry.size).to_s + '%'])
    end
end