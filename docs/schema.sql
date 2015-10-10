CREATE KEYSPACE bis WITH replication = {'class': 'SimpleStrategy', 'replication_factor': '1'};

create table bis.item(
  item_id int,
  name text,
  description text,
  keywords list<text>,
  category_path list<frozen<tuple<int,text>>>,
  brand_id int,
  brand_name text,
  flag_map map<text,boolean>,
  PRIMARY key (item_id)
) with comment = 'Catalog itens';

create table bis.item_attribute(
  item_id int,
  field_id int,
  name text,
  position int,
  group_id int,
  group_name text,
  group_position int,
  field_type_id int,
  field_type_name text,
  values list<frozen<tuple<int,text,text>>>,
  flag_map map<text,boolean>,
  primary key (item_id)
) with comment = 'Attributes for catalog itens';

create table variation(
  item_id int,
  variation_id int,
  name text,
  code_list map<text,frozen<list<text>>>,
  flag_map map<text,boolean>,
  primary key (variation_id)
)with comment = 'Item variation';

create table variation_by_item (
  item_id int,
  variation_id int,
  name text,
  code_list map<text,frozen<list<text>>>,
  flag_map map<text,boolean>,
  primary key (item_id, variation_id)
)with comment = 'Item variation, partitioned by item_id';

create table variation_attribute(
  variation_id int,
  item_id int,
  field_id int,
  name text,
  position int,
  group_id int,
  group_name text,
  group_position int,
  field_type_id int,
  field_type_name text,
  values list<frozen<tuple<int,text,text>>>,
  flag_map map<text,boolean>,
  primary key (variation_id)
)with comment = 'Dynamic variation attributes';

create table variation_attribute_by_item(
  variation_id int,
  item_id int,
  field_id int,
  name text,
  position int,
  group_id int,
  group_name text,
  group_position int,
  field_type_id int,
  field_type_name text,
  values list<frozen<tuple<int,text,text>>>,
  flag_map map<text,boolean>,
  primary key (item_id, variation_id)
)with comment = 'Dynamic variation attributes, partitioned by item_id';

create table variation_asset(
  variation_id int,
  item_id int,
  asset_id timeuuid,
  name text,
  description text,
  location text,
  type_id int,
  type_name text,
  type_description text,
  file_format_id int,
  file_format_name text,
  file_format_mimz text,
  file_format_extension text,
  flag_map map<text,boolean>,
  attribute_map map<text,text>,
  primary key (variation_id)
)with comment = 'Variation assets';

create table variation_asset_by_item(
  variation_id int,
  item_id int,
  asset_id timeuuid,
  name text,
  description text,
  location text,
  type_id int,
  type_name text,
  type_description text,
  file_format_id int,
  file_format_name text,
  file_format_mime text,
  file_format_extension text,
  flag_map map<text,boolean>,
  attribute_map map<text,text>,
  primary key (item_id, variation_id)
)with comment = 'Variation assets, partitioned by item_id';

create table offer(
  item_id  int,
  variation_id  int,
  offer_id  int,
  price_current  decimal,
  price_original  decimal,
  seller_id  text,
  seller_name  text,
  seller_offer_id  text,
  seller_status  boolean,
  flag_map map<text,boolean>,
  date_map map<text,timestamp>,
  primary key (offer_id)
)with comment = 'Offer of some item variation, on behalf of a seller. ';

create table offer_by_variation(
  item_id  int,
  variation_id  int,
  offer_id  int,
  price_current  decimal,
  price_original  decimal,
  seller_id  text,
  seller_name  text,
  seller_offer_id  text,
  seller_status  boolean,
  flag_map map<text,boolean>,
  date_map map<text,timestamp>,
  primary key (variation_id, offer_id)
)with comment = 'Offer of some item variation, on behalf of a seller. Partitioned by variation_id.';

create table offer_by_item(
  item_id  int,
  variation_id  int,
  offer_id  int,
  price_current  decimal,
  price_original  decimal,
  seller_id  text,
  seller_name  text,
  seller_offer_id  text,
  seller_status  boolean,
  flag_map map<text,boolean>,
  date_map map<text,timestamp>,
  primary key (item_id, variation_id, offer_id)
)with comment = 'Offer of some item variation, on behalf of a seller. Partitioned by item_id.';

