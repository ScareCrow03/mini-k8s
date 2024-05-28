import json
from datetime import datetime
import couchdb

class ImageProcessCommons:
    IMAGE_NAME = "imageName"

def main(args):
    current_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S.%f")[:-3]
    start_times = args.get("startTimes", [])
    start_times.append(current_time)

    response = args

    couchdb_url = args.get("COUCHDB_URL")
    if not couchdb_url:
        print("ExtractImageMetadata: missing COUCHDB_URL")
        return response
    
    couchdb_log_dbname = args.get("COUCHDB_LOGDB")
    if not couchdb_log_dbname:
        print("ExtractImageMetadata: missing COUCHDB_LOGDB")
        return response

    response["startTimes"] = start_times

    comm_times = args.get("commTimes", [])
    comm_times.append(0)
    response["commTimes"] = comm_times

    try:
        image_name = args.get(ImageProcessCommons.IMAGE_NAME)
        client = couchdb.Server(couchdb_url)
        db = client[couchdb_log_dbname]

        log = {
            "_id": str(datetime.utcnow().timestamp()).replace('.', ''),
            "img": image_name
        }
        db.save(log)
        
        response["log"] = log["_id"]

    except Exception as e:
        print(e)

    return response

def handle(param):
    """
    Parameters:
        args (dict): A dictionary containing the following keys:
            - COUCHDB_URL (str): 连接couchdb的url，注意，要在url中就固定好username和password，例如http://admin:123@localhost:5984
            - COUCHDB_DBNAME (str): 想要查看couchdb的数据库名
            - IMAGE_NAME (str): 想要查看的文档中的哪一个图片
            - IMAGE_DOCID (str): 想要查看数据库中的文档id
            - COUCHDB_LOGDB (str): 想要记录日志的数据库名
    """
    return main(param)