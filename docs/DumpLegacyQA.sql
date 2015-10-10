--Apelido do Banco: QA-Marketplace

select
isAvailable,
createdAt,
offerId,
priceCurrent,
priceImportTax,
priceOriginal,
productId,
quantity,
sellerId,
sellerName,
sellerOfferId,
sellerStatus,
skuId,
skuName,
statusSeller,
statusWalmart,
updatedAt
from (


  SELECT
    --top 10
    ss.SkuSellerId                                              offerId,
    s.IdSKU                                                     skuId,
    p.Id                                                        productId,
    cast(ss.SellerStockKeepingUnitId as varchar)                                 sellerOfferId,
    cast(ss.SellerId as VARCHAR)                                                 sellerId,
    se.IsActive                                                 sellerStatus,
    cast(s.PrecoCusto as int)                                   priceCurrent,
    cast(s.PrecoAnterior as int)                                priceOriginal,
    cast (1 as bit)                                             isAvailable,
    ss.IsActive                                                 statusSeller,
    ss.IsActive                                                 statusWalmart,
    (CAST(DATEDIFF(SECOND, '1969-12-31 21:00:00', CAST(ss.UpdateDate AS DATE)) AS BIGINT) * 1000) +
    DATEDIFF(MS, CAST(ss.UpdateDate AS DATE), ss.UpdateDate) AS createdAt,
    (CAST(DATEDIFF(SECOND, '1969-12-31 21:00:00', CAST(ss.UpdateDate AS DATE)) AS BIGINT) * 1000) +
    DATEDIFF(MS, CAST(ss.UpdateDate AS DATE), ss.UpdateDate) AS updatedAt,
    1                                                           priceImportTax,
    'skuname'                                                   skuName,
    se.Name                                                     sellerName,
    99999                                                       quantity

  FROM SkuSeller ss
    INNER JOIN Sku s ON s.IdSKU = ss.StockKeepingUnitId
    INNER JOIN Product p ON p.Id = s.IdProduto
    INNER JOIN Seller se ON se.SellerId = ss.SellerId
) as x


--COPY offer (available,created_at,offer_id,price_current,price_import_tax,price_original,product_id,quantity,seller_id,seller_name,seller_offer_id,seller_status,sku_id,sku_name,status_seller,status_walmart,updated_at) from 'allFromLegacy.csv';
--COPY offer_by_product (available,created_at,offer_id,price_current,price_import_tax,price_original,product_id,quantity,seller_id,seller_name,seller_offer_id,seller_status,sku_id,sku_name,status_seller,status_walmart,updated_at) from 'allFromLegacy.csv';

