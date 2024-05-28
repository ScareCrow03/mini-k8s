import couchdb
from PIL import Image
import json
import io
from datetime import datetime

def extract_image_metadata(args):
    current_time = datetime.now()

    print("ExtractImageMetadata invoked")

    response = args

    couchdb_url = args.get("COUCHDB_URL")
    if couchdb_url is None:
        print("ExtractImageMetadata: missing COUCHDB_URL")
        return response
    couchdb_dbname = args.get("COUCHDB_DBNAME")
    if couchdb_dbname is None:
        print("ExtractImageMetadata: missing COUCHDB_DBNAME")
        return response

    entry_time = current_time.strftime("%Y-%m-%d %H:%M:%S.%f")[:-3]
    response["startTimes"] = [entry_time]

    image_name = args.get("IMAGE_NAME")
    if image_name is None:
        print("ExtractImageMetadata: missing image name")
        return response

    try:
        server = couchdb.Server(couchdb_url)
        db = server[couchdb_dbname]
        
        doc_id = args.get("IMAGE_DOCID")
        if doc_id is None:
            print("ExtractImageMetadata: missing document ID")
            return response

        db_begin = datetime.now()
        doc = db.get(doc_id)
        if image_name not in doc['_attachments']:
            print(f"ExtractImageMetadata: missing attachment {image_name}")
            return response
        
        attachment = db.get_attachment(doc_id, image_name)
        db_finish = datetime.now()
        db_elapse_ms = (db_finish - db_begin).total_seconds() * 1000

        response["commTimes"] = [db_elapse_ms]

        image = Image.open(io.BytesIO(attachment.read()))
        metadata = {
            "format": image.format,
            "mode": image.mode,
            "size": image.size
        }

        response["IMAGE_NAME"] = image_name
        response["EXTRACTED_METADATA"] = metadata

    except Exception as e:
        print(e)

    return response

# Example usage
args = {
    "COUCHDB_URL": "http://admin:123@localhost:5984",
    "COUCHDB_DBNAME": "image",
    "imageName": "image.jpg",
    "imageDocId": "image1"
}

def handle(param):
    #param包括couchdb的url，username，password以及想要访问哪一个数据库db_name,以及数据库中的文档id imagedocid，最后是该文档包含的图像名字
    """
    Parameters:
        args (dict): A dictionary containing the following keys:
            - COUCHDB_URL (str): 连接couchdb的url，注意，要在url中就固定好username和password，例如http://admin:123@localhost:5984
            - COUCHDB_DBNAME (str): 想要查看couchdb的数据库名
            - IMAGE_NAME (str): 想要查看的文档中的哪一个图片
            - IMAGE_DOCID (str): 想要查看数据库中的文档id
    """
    resp = extract_image_metadata(param)
    json.dumps(resp, indent=2)
    return resp

