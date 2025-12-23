"""
向量检索服务 - 提供向量存储和检索功能
"""
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List, Optional
import os
from qdrant_client import QdrantClient
from qdrant_client.models import Distance, VectorParams, PointStruct
import httpx

app = FastAPI(title="Vector Service", version="1.0.0")

# 初始化Qdrant客户端
qdrant_host = os.getenv("QDRANT_HOST", "localhost")
qdrant_port = int(os.getenv("QDRANT_PORT", "6333"))
qdrant_client = QdrantClient(host=qdrant_host, port=qdrant_port)

# Embedding服务URL
embedding_service_url = os.getenv("EMBEDDING_SERVICE_URL", "http://localhost:8002")

# 集合名称
COLLECTION_NAME = "knowledge_base"
# 向量维度：本地模型 paraphrase-multilingual-MiniLM-L12-v2 的维度是 384
# 如果使用其他模型，需要相应调整
VECTOR_SIZE = int(os.getenv("VECTOR_SIZE", "384"))


class SearchRequest(BaseModel):
    query: str
    limit: int = 5
    knowledge_base_id: Optional[str] = None


class SearchResult(BaseModel):
    id: str
    score: float
    payload: dict


class SearchResponse(BaseModel):
    results: List[SearchResult]


class StoreRequest(BaseModel):
    id: str
    vector: List[float]
    payload: dict


@app.on_event("startup")
async def startup():
    """启动时确保集合存在"""
    try:
        collections = qdrant_client.get_collections().collections
        collection_names = [col.name for col in collections]
        
        if COLLECTION_NAME not in collection_names:
            qdrant_client.create_collection(
                collection_name=COLLECTION_NAME,
                vectors_config=VectorParams(
                    size=VECTOR_SIZE,
                    distance=Distance.COSINE
                )
            )
            print(f"创建集合: {COLLECTION_NAME}")
    except Exception as e:
        print(f"初始化集合失败: {e}")


@app.post("/search", response_model=SearchResponse)
async def search(request: SearchRequest):
    """
    向量检索
    """
    try:
        # 获取查询文本的向量
        async with httpx.AsyncClient() as client:
            embed_response = await client.post(
                f"{embedding_service_url}/embed",
                json={"text": request.query},
                timeout=30.0
            )
            embed_response.raise_for_status()
            embedding = embed_response.json()["embedding"]
        
        # 构建过滤条件
        filter_condition = None
        if request.knowledge_base_id:
            filter_condition = {
                "must": [
                    {
                        "key": "knowledge_base_id",
                        "match": {"value": request.knowledge_base_id}
                    }
                ]
            }
        
        # 执行检索
        search_results = qdrant_client.search(
            collection_name=COLLECTION_NAME,
            query_vector=embedding,
            limit=request.limit,
            query_filter=filter_condition
        )
        
        results = [
            SearchResult(
                id=result.id,
                score=result.score,
                payload=result.payload
            )
            for result in search_results
        ]
        
        return SearchResponse(results=results)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"检索失败: {str(e)}")


@app.post("/store")
async def store(request: StoreRequest):
    """
    存储向量
    """
    try:
        if len(request.vector) != VECTOR_SIZE:
            raise HTTPException(
                status_code=400,
                detail=f"向量维度不匹配，期望{VECTOR_SIZE}，实际{len(request.vector)}"
            )
        
        point = PointStruct(
            id=request.id,
            vector=request.vector,
            payload=request.payload
        )
        
        qdrant_client.upsert(
            collection_name=COLLECTION_NAME,
            points=[point]
        )
        
        return {"status": "ok", "id": request.id}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"存储失败: {str(e)}")


@app.delete("/delete/{point_id}")
async def delete_point(point_id: str):
    """
    删除向量点
    """
    try:
        qdrant_client.delete(
            collection_name=COLLECTION_NAME,
            points_selector=[point_id]
        )
        return {"status": "ok", "id": point_id}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"删除失败: {str(e)}")


@app.get("/health")
async def health():
    """健康检查"""
    try:
        collections = qdrant_client.get_collections()
        return {"status": "ok", "collections": len(collections.collections)}
    except Exception as e:
        return {"status": "error", "error": str(e)}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)

