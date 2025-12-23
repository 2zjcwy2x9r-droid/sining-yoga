"""
Embedding服务 - 提供文本向量化功能（使用本地模型）
"""
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import os
from typing import List
from sentence_transformers import SentenceTransformer

app = FastAPI(title="Embedding Service", version="1.0.0")

# 初始化本地embedding模型
embedding_model_name = os.getenv("EMBEDDING_MODEL", "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2")
print(f"正在加载embedding模型: {embedding_model_name}...")
try:
    model = SentenceTransformer(embedding_model_name)
    print(f"✓ 模型加载成功: {embedding_model_name}")
except Exception as e:
    print(f"✗ 模型加载失败: {e}")
    raise ValueError(f"无法加载embedding模型 {embedding_model_name}: {str(e)}")


class EmbedRequest(BaseModel):
    text: str


class EmbedResponse(BaseModel):
    embedding: List[float]
    model: str


@app.post("/embed", response_model=EmbedResponse)
async def embed_text(request: EmbedRequest):
    """
    将文本转换为向量（使用本地模型）
    """
    try:
        # 使用本地模型生成embedding
        embedding = model.encode(request.text, normalize_embeddings=True).tolist()
        
        return EmbedResponse(
            embedding=embedding,
            model=embedding_model_name
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"向量化失败: {str(e)}")


@app.get("/health")
async def health():
    """健康检查"""
    return {"status": "ok", "model": embedding_model_name}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
