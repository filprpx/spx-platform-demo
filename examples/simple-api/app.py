from fastapi import FastAPI


app = FastAPI(title="SPX Demo API")


@app.get("/")
def health():
    return {"application": "SPX demo API", "status": "ok"}
