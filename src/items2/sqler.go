package main

// Handles transforming of field maps into SQL queries.

// TODO(amit): Needs refactoring. This is no longer an SQL generator.

import (
	"bouncer"
	"strings"
)

// ----- SQLER TYPE ------------------------------------------------------------

// Takes parsed entries from the XMLs and generates SQL queries for them.
// The time argument is used for creating timestamps. It should hold the time
// the data was published, in seconds since 1/1/1970 (Unix time).
type sqler func(data []map[string]string, time int64) []byte

// All available sqlers.
var sqlers = map[string]sqler{
	"stores": storesSqler,
	"prices": pricesSqler,
	"promos": promosSqler,
}

// ----- CONCRETE SQLERS -------------------------------------------------------

// Insert commands should be performed in batches, since there is a limit
// on the maximal insert size in SQLite.
const batchSize = 500

// Creates SQL statements for stores.
func storesSqler(data []map[string]string, time int64) []byte {
	data = escapeQuotes(data)

	// Get store-ids.
	ss := make([]*bouncer.Store, len(data))
	for i, d := range data {
		ss[i] = &bouncer.Store{
			d["chain_id"],
			d["subchain_id"],
			d["store_id"],
		}
	}
	sids := bouncer.MakeStoreIds(ss)

	// Report store-metas.
	metas := make([]*bouncer.StoreMeta, len(data))
	for i, d := range data {
		metas[i] = &bouncer.StoreMeta{
			time,
			sids[i],
			d["bikoret_no"],
			d["store_type"],
			d["chain_name"],
			d["subchain_name"],
			d["store_name"],
			d["address"],
			d["city"],
			d["zip_code"],
			d["last_update_date"],
			d["last_update_time"],
		}
	}

	bouncer.ReportStoreMetas(metas)

	return nil
}

// Creates SQL statements for prices.
func pricesSqler(data []map[string]string, time int64) []byte {
	data = escapeQuotes(data)

	// Report stores (just to get ids).
	ss := make([]*bouncer.Store, len(data))
	for i, d := range data {
		ss[i] = &bouncer.Store{
			d["chain_id"],
			d["subchain_id"],
			d["store_id"],
		}
	}
	sids := bouncer.MakeStoreIds(ss)

	// Report items.
	is := make([]*bouncer.Item, len(data))
	for i, d := range data {
		is[i] = &bouncer.Item{d["item_type"], d["item_code"], d["chain_id"]}
		if is[i].ItemType != `"0"` {
			is[i].ItemType = `"1"`
			is[i].ChainId = ""
		}
	}
	ids := bouncer.MakeItemIds(is)

	// Report item-metas.
	metas := make([]*bouncer.ItemMeta, len(data))
	for i, d := range data {
		metas[i] = &bouncer.ItemMeta{
			time,
			ids[i],
			sids[i],
			d["update_time"],
			d["item_name"],
			d["manufacturer_item_description"],
			d["unit_quantity"],
			d["is_weighted"],
			d["quantity_in_package"],
			d["allow_discount"],
			d["item_status"],
		}
	}

	bouncer.ReportItemMetas(metas)

	// Report prices.
	prices := make([]*bouncer.Price, len(data))
	for i, d := range data {
		prices[i] = &bouncer.Price{
			time,
			ids[i],
			sids[i],
			d["price"],
			d["unit_of_measure_price"],
			d["unit_of_measure"],
			d["quantity"],
		}
	}

	bouncer.ReportPrices(prices)

	return nil
}

// Creates SQL statements for promos.
func promosSqler(data []map[string]string, time int64) []byte {
	if len(data) == 0 {
		return nil
	}

	data = escapeQuotes(data)

	// Get store id.
	sid := bouncer.MakeStoreIds([]*bouncer.Store{&bouncer.Store{
		data[0]["chain_id"],
		data[0]["subchain_id"],
		data[0]["store_id"],
	}})[0]

	// Create promos.
	promos := make([]*bouncer.Promo, len(data))
	for i, d := range data {
		promos[i] = &bouncer.Promo{
			time,
			d["chain_id"],
			d["promotion_id"],
			d["promotion_description"],
			d["promotion_start_date"],
			d["promotion_start_hour"],
			d["promotion_end_date"],
			d["promotion_end_hour"],
			d["reward_type"],
			d["allow_multiple_discounts"],
			d["min_qty"],
			d["max_qty"],
			d["discount_rate"],
			d["discount_type"],
			d["min_purchase_amnt"],
			d["min_no_of_item_offered"],
			d["price_update_date"],
			d["discounted_price"],
			d["discounted_price_per_mida"],
			d["additional_is_coupn"],
			d["additional_gift_count"],
			d["additional_is_total"],
			d["additional_min_basket_amount"],
			d["remarks"],
			sid,
			nil,
			nil,
		}
	}

	// Create item ids.
	for i, d := range data {
		// Get rid of quotes (generated by escape quotes)
		d["item_code"] = d["item_code"][1 : len(d["item_code"])-1]
		d["item_type"] = d["item_type"][1 : len(d["item_type"])-1]
		d["is_gift_item"] = d["is_gift_item"][1 : len(d["is_gift_item"])-1]

		// Get repeated fields.
		codes := strings.Split(d["item_code"], ";")
		types := strings.Split(d["item_type"], ";")
		gifts := strings.Split(d["is_gift_item"], ";")

		// Check lengths are all equal.
		if len(codes) != len(types) {
			// TODO(amit): Return an error.
			pe("Promo ignored promo due to mismatching lengths:", len(codes),
				len(types))
			continue
		}

		// Generate items.
		items := make([]*bouncer.Item, len(codes))
		for j := range codes {
			items[j] = &bouncer.Item{types[j], codes[j], promos[i].ChainId}
			if items[j].ItemType != `"0"` {
				items[j].ItemType = `"1"`
				items[j].ChainId = ""
			}
		}
		promos[i].ItemIds = bouncer.MakeItemIds(items)

		// Generate gift items.
		promos[i].GiftItems = make([]string, len(codes))
		if len(gifts) == len(codes) {
			for j := range gifts {
				promos[i].GiftItems[j] = gifts[j]
			}
		}
	}

	bouncer.ReportPromos(promos)

	return nil
}

// ----- OTHER HELPERS ---------------------------------------------------------

// Escapes quotation characters in parsed data. Input data is unchanged.
func escapeQuotes(maps []map[string]string) []map[string]string {
	result := make([]map[string]string, len(maps))
	for i := range maps {
		result[i] = map[string]string{}
		for k, v := range maps[i] {
			result[i][k] = "\"" + strings.Replace(v, "\"", "\"\"", -1) + "\""
		}
	}
	return result
}

