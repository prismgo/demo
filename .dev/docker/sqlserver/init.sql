IF DB_ID(N'$(DatabaseName)') IS NULL
BEGIN
    EXEC(N'CREATE DATABASE ' + QUOTENAME(N'$(DatabaseName)'));
END;
GO
