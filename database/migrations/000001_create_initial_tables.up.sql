-- tabela de clientes
CREATE TABLE clients (
    id BIGSERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    status BOOLEAN DEFAULT TRUE,
    data_cadastro TIMESTAMPTZ NOT NULL
);

-- tabela de produtos
CREATE TABLE client_product (
    id BIGSERIAL PRIMARY KEY,
    nome_plano VARCHAR(255) UNIQUE NOT NULL,
    preco_centavos BIGINT NOT NULL,
    ativo BOOLEAN DEFAULT TRUE
);