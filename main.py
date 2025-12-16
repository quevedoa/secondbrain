from fastapi import FastAPI
from pydantic import BaseModel
from ollama import chat
from ollama import ChatResponse
import chromadb
import datetime

CHROMA_DB_PATH="./chromadb"

app = FastAPI()
chroma_db_client = chromadb.PersistentClient(path=CHROMA_DB_PATH)

collection = chroma_db_client.get_or_create_collection(name="secondbrainchroma")

class Note(BaseModel):
    content: str

class NoteQuery(BaseModel):
    query: str
    num_results: int = 10

@app.post("/addNote", status_code=201)
def add_note(note: Note):
    current_time = datetime.datetime.now()
    note_path = f"./notes/{str(current_time)}.txt"
    with open(note_path, "w") as f:
        f.write(note.content)
    collection.add(
        ids=[str(current_time)],
        documents=[note.content]
    )
    return {"filename": note_path}

@app.post("/queryNote", status_code=201)
def query_note(note_query: NoteQuery):
    results = collection.query(
        query_texts=[note_query.query],
        n_results=note_query.num_results
    )

    prompt = f"{note_query.query}. Answer this by using the following context, make sure to keep your response as short as possible and only summarize what answers my question given the context.:"
    for query_results in results["documents"]:
        prompt += prompt + "\n".join(query_results)

    response = chat(model='llama3', messages=[
        {
            'role': 'user',
            'content': prompt,
        },
    ])
    return {"response": response.message.content}




    
    