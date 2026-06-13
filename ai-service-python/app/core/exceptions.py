from fastapi import HTTPException, status


class AppException(HTTPException):
    def __init__(self, code: int, message: str, status_code: int = status.HTTP_400_BAD_REQUEST):
        super().__init__(status_code=status_code, detail={"code": code, "message": message})


class NotFoundError(AppException):
    def __init__(self, message: str = "资源不存在"):
        super().__init__(code=40401, message=message, status_code=status.HTTP_404_NOT_FOUND)


class UnauthorizedError(AppException):
    def __init__(self, message: str = "未登录或登录已过期"):
        super().__init__(code=40101, message=message, status_code=status.HTTP_401_UNAUTHORIZED)
