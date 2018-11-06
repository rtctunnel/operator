package operator

// A ChannelCollection is a collection of channels for a given topic.
type ChannelCollection map[string]map[chan string]struct{}

// Add adds a channel to the collection.
func (cc ChannelCollection) Add(topic string, c chan string) {
	cs, ok := cc[topic]
	if !ok {
		cs = make(map[chan string]struct{})
		cc[topic] = cs
	}
	cs[c] = struct{}{}
}

// List lists all the channels for the given topic
func (cc ChannelCollection) List(topic string) []chan string {
	var result []chan string
	cs, ok := cc[topic]
	if ok {
		for c := range cs {
			result = append(result, c)
		}
	}
	return result
}

// Remove removes a channel from the collection.
func (cc ChannelCollection) Remove(topic string, c chan string) {
	cs, ok := cc[topic]
	if ok {
		delete(cs, c)
		if len(cs) == 0 {
			delete(cc, topic)
		}
	}
}
