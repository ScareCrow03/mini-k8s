FROM python:3.8-slim

COPY . /app

WORKDIR /app

RUN pip3 config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple
RUN pip install -r requirements.txt

EXPOSE 10000

CMD ["python", "baseserver.py"]