

select distinct p.IdProduto, ss.SkuSellerId
from produto p
  inner join sku s on s.IdProduto = p.IdProduto
  inner join SkuSeller ss on ss.StockKeepingUnitId = s.IdSKU

where
  p.FlagAtiva = 1 AND
  s.FlagAtiva = 1 AND
  ss.IsActive = 1 and
  ss.SellerId not in ('1','0')