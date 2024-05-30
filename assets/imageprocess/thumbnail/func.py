import couchdb
import json
import subprocess
from datetime import datetime
import wand
import wand.image

class Thumbnail:
    LAUNCH_TIME = datetime.now()

    MAX_WIDTH = 250
    MAX_HEIGHT = 250

    @staticmethod
    def main(args):
        entry_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S.%f")[:-3]
        start_times = args.get("startTimes", [])
        start_times.append(entry_time)
        response = args
        response["startTimes"] = start_times

        couchdb_url = args.get("COUCHDB_URL")
        couchdb_dbname = args.get("COUCHDB_DBNAME")
        image_docid = args.get("IMAGE_DOCID")

        if not all([couchdb_url, couchdb_dbname]):
            print("Thumbnail: missing CouchDB configuration")
            return response

        try:
            image_name = args.get("IMAGE_NAME")

            server = couchdb.Server(couchdb_url)
            # server.resource.credentials = (couchdb_username, couchdb_password)
            db = server[couchdb_dbname]

            db_begin = datetime.now()
            doc = db.get_attachment(image_docid, image_name)
            db_finish = datetime.now()
            db_elapse_ms = (db_finish - db_begin).total_seconds() * 1000

            comm_times = args.get("commTimes", [])
            comm_times.append(db_elapse_ms)
            response["commTimes"] = comm_times

            with open(image_name, "wb") as image_stream:
                image_stream.write(doc.read())

            width = args["EXTRACTED_METADATA"]["width"]
            height =  args["EXTRACTED_METADATA"]["height"]

            scaling_factor = min(Thumbnail.MAX_HEIGHT / height, Thumbnail.MAX_WIDTH / width)
            width = int(width * scaling_factor)
            height = int(height * scaling_factor)

            thumbnail_name = "thumbnail-" + image_name
            with wand.image.Image(filename=image_name) as img:
                img.resize(width, height)
                img.save(filename=thumbnail_name)

            with open(thumbnail_name, "rb") as thumbnail_stream:
                doc = db.get(image_docid)
                db.put_attachment(doc, thumbnail_stream, filename=thumbnail_name)

            response["thumbnail"] = thumbnail_name

        except Exception as e:
            print("Error:", e)

        end_time = datetime.now()
        execution_time = (end_time - Thumbnail.LAUNCH_TIME).total_seconds()
        response["executionTime"] = execution_time

        return response


def handle(param):
    """
    Parameters:
        args (dict): A dictionary containing the following keys:
            - COUCHDB_URL (str): 连接couchdb的url，注意，要在url中就固定好username和password，例如http://admin:123@localhost:5984
            - COUCHDB_DBNAME (str): 想要查看couchdb的数据库名
            - IMAGE_NAME (str): 想要查看的文档中的哪一个图片
            - IMAGE_DOCID (str): 想要查看数据库中的文档id
    """
    return Thumbnail.main(param)


# 示例用法
# json_str = '''{
#     "extractedMetadata": {
#         "creationTime": "2019:10:15 14:03:39",
#         "dimensions": {
#             "height": 3968,
#             "width": 2976
#         },
#         "exifMake": "HUAWEI",
#         "exifModel": "ALP-AL00",
#         "fileSize": "2.372MB",
#         "format": "image/jpeg",
#         "geo": {
#             "latitude": {
#                 "D": 31,
#                 "Direction": "N",
#                 "M": 1,
#                 "S": 27
#             },
#             "longitude": {
#                 "D": 121,
#                 "Direction": "E",
#                 "M": 26,
#                 "S": 15
#             }
#         }
#     },
#     "imageName": "image.jpg"
# }'''

# json_args = json.loads(json_str)
# Thumbnail.main(json_args)
