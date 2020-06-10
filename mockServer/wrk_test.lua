--wrk.method = "POST"
--wrk.headers["Content-Type"] = "application/x-www-form-urlencoded"
--wrk.body = "token=ABC&age=20"

-- Load URL paths from the file
function load_url_paths_from_file(file)
    lines = {}

    -- Check if the file exists
    -- Resource: http://stackoverflow.com/a/4991602/325852
    local f=io.open(file,"r")
    if f~=nil then
        io.close(f)
    else
        -- Return the empty array
        return lines
    end

    -- If the file exists loop through all its lines
    -- and add them into the lines array
    for line in io.lines(file) do
        if not (line == '') then
            lines[#lines + 1] = line
        end
    end

    return lines
end

-- Load URL paths from file
paths = load_url_paths_from_file("./test_uri_comparison.txt")

count = 1

request = function()
    if count > #paths then
        count = 1
    end

    url_path = paths[count]
    count = count + 1

    if count >= 6 then
        body = paths[count]
        count = count + 1
        return wrk.format('POST', url_path, {["Content-Type"] = "application/x-www-form-urlencoded"}, body)
    end
    return wrk.format('GET', url_path, {["Content-Type"] = "application/x-www-form-urlencoded"}, nil)
end