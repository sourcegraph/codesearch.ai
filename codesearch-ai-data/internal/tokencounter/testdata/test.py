def gz(path, *args, **kwargs):
    """
    Loads a dataframe from a gz file.
    :param path: path or location of the file. Must be string dataType
    :param args: custom argument to be passed to the internal function
    :param kwargs: custom keyword arguments to be passed to the internal function
    :return: Spark Dataframe
    """
    file, file_name = prepare_path(path, "gz")

    import gzip
    import shutil

    with gzip.open(file, "rb") as f_in:
        print(f_in)
        with open("file.txt", "wb") as f_out:
            shutil.copyfileobj(f_in, f_out)

    print(file, file_name)
