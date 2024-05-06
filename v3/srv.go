package dnsutils

import (
	"math/rand"
	"net"
	"sort"
)

// OrderedSRV returns a count of the results and a map keyed on the order they should be used.
// This based on the records' priority and randomised selection based on their relative weighting.
// The function's inputs are the same as those for net.LookupSRV
// To use in the correct order:
//
// count, orderedSRV, err := OrderedSRV(service, proto, name)
//
//	for  _, srv := range orderedSRV {
//	  // Do something such as dial this SRV. If fails move on the the next or break if it succeeds.
//	}
func OrderedSRV(service, proto, name string) (int, []*net.SRV, error) {
	_, addrs, err := net.LookupSRV(service, proto, name)
	if err != nil {
		return 0, nil, err
	}
	index, osrv := OrderSRV(addrs)
	return index, osrv, nil
}

func OrderSRV(addrs []*net.SRV) (int, []*net.SRV) {
	// No need to process less than two records
	if cnt := len(addrs); cnt < 2 {
		return cnt, addrs
	}

	// Initialise the ordered map
	osrv := make([]*net.SRV, 0, len(addrs))

	prioMap := make(map[int][]*net.SRV, 0)
	for _, srv := range addrs {
		prioMap[int(srv.Priority)] = append(prioMap[int(srv.Priority)], srv)
	}

	priorities := make([]int, 0, len(prioMap))
	for p := range prioMap {
		priorities = append(priorities, p)
	}

	var count int
	sort.Ints(priorities)
	for _, p := range priorities {
		tos := weightedOrder(prioMap[p])
		for _, s := range tos {
			count++
			osrv = append(osrv, s)
		}
	}
	return count, osrv
}

func weightedOrder(srvs []*net.SRV) []*net.SRV {
	// No need to process less than two records
	if len(srvs) < 2 {
		return srvs
	}

	// Get the total weight
	var tw int
	for _, s := range srvs {
		tw += int(s.Weight)
	}

	// Initialise the ordered map
	osrv := make([]*net.SRV, 0, len(srvs))

	// Whilst there are still entries to be ordered
	l := len(srvs)
	for l > 0 {
		i := rand.Intn(l)
		s := srvs[i]
		var rw int
		if tw > 0 {
			// Greater the weight the more likely this will be zero or less
			rw = rand.Intn(tw) - int(s.Weight)
		}
		if rw <= 0 {
			// Put entry in position
			osrv = append(osrv, s)
			if len(srvs) > 1 {
				// Remove the entry from the source slice by swapping with the last entry and truncating
				srvs[len(srvs)-1], srvs[i] = srvs[i], srvs[len(srvs)-1]
				srvs = srvs[:len(srvs)-1]
				l = len(srvs)
			} else {
				l = 0
			}
			tw = tw - int(s.Weight)
		}
	}
	return osrv
}
