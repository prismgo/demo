IF DB_ID(N'$(DatabaseName)') IS NULL
BEGIN
    DECLARE @CreateDatabaseSQL nvarchar(max);
    SET @CreateDatabaseSQL = N'CREATE DATABASE ' + QUOTENAME(N'$(DatabaseName)');
    EXEC sys.sp_executesql @CreateDatabaseSQL;
END;
GO
